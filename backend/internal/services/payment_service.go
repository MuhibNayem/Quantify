package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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
	ValidateSSLCommerzIPN(formValue map[string][]string) (bool, error)
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
	jsonPayload, _ := json.Marshal(payload)

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

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if accessToken, ok := result["id_token"].(string); ok {
		return accessToken, nil
	}
	return "", fmt.Errorf("failed to get bKash access token: %v", result)
}

func (s *paymentService) CreatePayment(paymentMethod string, total float64, tranID, currency, cusName, cusEmail, cusPhone string) (string, error) {
	// Create a new transaction with a "pending" status
	transaction := &domain.Transaction{
		OrderID:       tranID,
		Amount:        int64(total * 100), // Store amount in smallest currency unit
		Currency:      currency,
		PaymentMethod: paymentMethod,
		Status:        "pending",
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
	jsonPayload, _ := json.Marshal(payload)

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

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

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
	jsonExecutePayload, _ := json.Marshal(executePayload)

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

	executeBody, _ := ioutil.ReadAll(executeResp.Body)
	var executeResult map[string]interface{}
	json.Unmarshal(executeBody, &executeResult)

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

func (s *paymentService) ValidateSSLCommerzIPN(formValue map[string][]string) (bool, error) {
	// This is a simplified validation. A real implementation would involve
	// verifying the hash and other parameters as per SSLCommerz documentation.
	status := formValue["status"]
	if len(status) > 0 {
		tranID := formValue["tran_id"][0]
		transaction, err := s.paymentRepo.GetTransactionByTranID(tranID)
		if err != nil {
			return false, fmt.Errorf("failed to get transaction: %w", err)
		}

		if status[0] == "VALID" {
			transaction.Status = "completed"
			transaction.GatewayTransactionID = formValue["val_id"][0]
			if err := s.paymentRepo.UpdateTransaction(transaction); err != nil {
				return false, fmt.Errorf("failed to update transaction: %w", err)
			}
			return true, nil
		} else {
			transaction.Status = "failed"
			if err := s.paymentRepo.UpdateTransaction(transaction); err != nil {
				return false, fmt.Errorf("failed to update transaction: %w", err)
			}
			return false, fmt.Errorf("SSLCommerz IPN validation failed")
		}
	}
	return false, fmt.Errorf("invalid IPN data")
}
