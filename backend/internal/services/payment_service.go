package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
)

const (
	PaymentMethodBkash = "bkash"
	PaymentMethodCard  = "card"
	PaymentMethodCash  = "cash"
)

type PaymentService interface {
	CreatePayment(paymentMethod string, total float64, tranID, currency, cusName, cusEmail, cusPhone string) (string, error)
	HandleBkashCallback(paymentID, status string) (string, error)
}

type paymentService struct {
	cfg         *config.Config
	paymentRepo repository.PaymentRepository
}

func NewPaymentService(cfg *config.Config, paymentRepo repository.PaymentRepository) PaymentService {
	return &paymentService{cfg: cfg, paymentRepo: paymentRepo}
}

func (s *paymentService) getBkashAccessToken() (string, error) {
	url := s.cfg.BkashBaseURL + "/token/grant"
	payload := map[string]string{
		"app_key":    s.cfg.BkashAPIKey,
		"app_secret": s.cfg.BkashAPISecret,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal bKash token payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create bKash token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("username", s.cfg.BkashUsername)
	req.Header.Set("password", s.cfg.BkashPassword)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send bKash token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read bKash token response: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse bKash token response: %w", err)
	}

	if accessToken, ok := result["id_token"].(string); ok {
		return accessToken, nil
	}
	return "", fmt.Errorf("failed to get bKash access token: %v", result)
}

func (s *paymentService) CreatePayment(paymentMethod string, total float64, tranID, currency, cusName, cusEmail, cusPhone string) (string, error) {
	// Create a new transaction with a "pending" status
	transaction := &domain.Transaction{
		OrderID:              tranID,
		Amount:               int64(total * 100), // Store amount in smallest currency unit
		Currency:             currency,
		PaymentMethod:        paymentMethod,
		Status:               "pending",
		GatewayTransactionID: uuid.NewString(),
	}
	if err := s.paymentRepo.CreateTransaction(transaction); err != nil {
		return "", fmt.Errorf("failed to create transaction record: %w", err)
	}

	switch paymentMethod {
	case PaymentMethodBkash:
		return s.createBkashPayment(total, tranID, currency)
	case PaymentMethodCard:
		return s.createSSLCommerzCardPayment(total, tranID, currency, cusName, cusEmail, cusPhone)
	case PaymentMethodCash:
		return s.createCashPayment(total, tranID, currency)
	default:
		return "", fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}
}

func (s *paymentService) createCashPayment(total float64, tranID, currency string) (string, error) {
	// For cash payments, we can consider the payment as successful immediately.
	transaction, err := s.paymentRepo.GetTransactionByTranID(tranID)
	if err != nil {
		return "", fmt.Errorf("failed to get transaction: %w", err)
	}
	transaction.Status = "completed"
	transaction.GatewayTransactionID = fmt.Sprintf("cash-%s", tranID)
	if err := s.paymentRepo.UpdateTransaction(transaction); err != nil {
		return "", fmt.Errorf("failed to update transaction: %w", err)
	}
	return s.cfg.SSLCommerzSuccessURL, nil
}

func (s *paymentService) createBkashPayment(total float64, tranID, currency string) (string, error) {
	accessToken, err := s.getBkashAccessToken()
	if err != nil {
		return "", err
	}

	url := s.cfg.BkashBaseURL + "/create"
	payload := map[string]interface{}{
		"mode":            "0011", // Checkout mode
		"payerReference":  "12345",
		"callbackURL":     s.cfg.SSLCommerzSuccessURL, // Using SSLCommerz success URL for bKash callback for now
		"amount":          fmt.Sprintf("%.2f", total),
		"currency":        currency,
		"intent":          "sale",
		"merchantInvoice": tranID,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal bKash payment payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create bKash payment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", accessToken)
	req.Header.Set("X-App-Key", s.cfg.BkashAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send bKash payment request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read bKash payment response: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse bKash payment response: %w", err)
	}

	if bkashURL, ok := result["bkashURL"].(string); ok {
		return bkashURL, nil
	}
	return "", fmt.Errorf("failed to create bKash payment: %v", result)
}

func (s *paymentService) createSSLCommerzCardPayment(total float64, tranID, currency, cusName, cusEmail, cusPhone string) (string, error) {
	// Manually construct SSLCommerz request for card payment
	postBody := map[string]string{
		"store_id":         s.cfg.SSLCommerzStoreID,
		"store_passwd":     s.cfg.SSLCommerzStorePassword,
		"total_amount":     fmt.Sprintf("%.2f", total),
		"currency":         currency,
		"tran_id":          tranID,
		"success_url":      s.cfg.SSLCommerzSuccessURL,
		"fail_url":         s.cfg.SSLCommerzFailURL,
		"cancel_url":       s.cfg.SSLCommerzCancelURL,
		"ipn_url":          s.cfg.SSLCommerzIPNURL,
		"cus_name":         cusName,
		"cus_email":        cusEmail,
		"cus_phone":        cusPhone,
		"product_name":     "Inventory Product",
		"product_category": "General",
		"product_profile":  "general",
		"shipping_method":  "NO",
		"num_of_item":      "1",
		"value_a":          "ref001", // Example additional parameter
	}

	jsonData, err := json.Marshal(postBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SSLCommerz request: %w", err)
	}

	resp, err := http.Post(s.cfg.SSLCommerzAPIHost+"/gwprocess/v4/api.php", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send SSLCommerz request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read SSLCommerz response: %w", err)
	}

	var sslCommerzResponse struct {
		Status         string `json:"status"`
		FailedReason   string `json:"failedreason"`
		GatewayPageURL string `json:"GatewayPageURL"`
	}
	if err := json.Unmarshal(body, &sslCommerzResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal SSLCommerz response: %w", err)
	}

	if sslCommerzResponse.Status != "SUCCESS" {
		return "", fmt.Errorf("SSLCommerz payment initiation failed: %s", sslCommerzResponse.FailedReason)
	}

	return sslCommerzResponse.GatewayPageURL, nil
}

func (s *paymentService) HandleBkashCallback(paymentID, status string) (string, error) {
	if status != "success" {
		return s.cfg.SSLCommerzFailURL, nil
	}

	accessToken, err := s.getBkashAccessToken()
	if err != nil {
		return "", err
	}

	// Execute bKash payment
	executeURL := s.cfg.BkashBaseURL + "/execute"
	executePayload := map[string]string{"paymentID": paymentID}
	jsonExecutePayload, err := json.Marshal(executePayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal bKash execute payload: %w", err)
	}

	executeReq, err := http.NewRequest("POST", executeURL, bytes.NewBuffer(jsonExecutePayload))
	if err != nil {
		return "", fmt.Errorf("failed to create bKash execute request: %w", err)
	}
	executeReq.Header.Set("Content-Type", "application/json")
	executeReq.Header.Set("Authorization", accessToken)
	executeReq.Header.Set("X-App-Key", s.cfg.BkashAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	executeResp, err := client.Do(executeReq)
	if err != nil {
		return "", fmt.Errorf("failed to send bKash execute request: %w", err)
	}
	defer executeResp.Body.Close()

	executeBody, err := ioutil.ReadAll(executeResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read bKash execute response: %w", err)
	}
	var executeResult map[string]interface{}
	if err := json.Unmarshal(executeBody, &executeResult); err != nil {
		return "", fmt.Errorf("failed to parse bKash execute response: %w", err)
	}

	if transactionStatus, ok := executeResult["transactionStatus"].(string); ok && transactionStatus == "Completed" {
		// Payment successful, update transaction in DB
		tranID, ok := executeResult["merchantInvoice"].(string)
		if !ok {
			return s.cfg.SSLCommerzFailURL, fmt.Errorf("merchantInvoice not found in bKash execute result")
		}
		transaction, err := s.paymentRepo.GetTransactionByTranID(tranID)
		if err != nil {
			return s.cfg.SSLCommerzFailURL, fmt.Errorf("failed to get transaction: %w", err)
		}
		transaction.Status = "completed"
		transaction.GatewayTransactionID = paymentID
		if err := s.paymentRepo.UpdateTransaction(transaction); err != nil {
			return s.cfg.SSLCommerzFailURL, fmt.Errorf("failed to update transaction: %w", err)
		}
		return s.cfg.SSLCommerzSuccessURL, nil
	}

	return s.cfg.SSLCommerzFailURL, fmt.Errorf("bKash payment execution failed: %v", executeResult)
}

func ValidateSSLCommerzIPN(params map[string]string, storePass string, verify map[string]string) error {
	sign, ok := verify["verify_sign"]
	if !ok || sign == "" {
		return errors.New("missing signature")
	}
	keyList := verify["verify_key"]
	if keyList == "" {
		return errors.New("missing verify_key")
	}
	var parts []string
	for _, k := range strings.Split(keyList, ",") {
		k = strings.TrimSpace(k)
		if v, ok := params[k]; ok && v != "" {
			parts = append(parts, k+"="+v)
		}
	}
	sort.Strings(parts)
	hashSeq := strings.Join(parts, ",")
	mac := hmac.New(sha512.New, []byte(storePass))
	mac.Write([]byte(hashSeq))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sign), []byte(expected)) {
		return errors.New("signature mismatch")
	}
	if params["status"] == "" || !strings.HasPrefix(params["status"], "VALID") {
		return errors.New("invalid transaction status")
	}
	return nil
}

func ValidateSSLCommerzFromValues(form url.Values, storePass string) error {
	params := make(map[string]string, len(form))
	for k, v := range form {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	verify := map[string]string{
		"verify_sign": form.Get("verify_sign"),
		"verify_key":  form.Get("verify_key"),
	}
	return ValidateSSLCommerzIPN(params, storePass, verify)
}
