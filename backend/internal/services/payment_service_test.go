package services

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSSLCommerzIPN_RejectsMissingSignature(t *testing.T) {
	err := ValidateSSLCommerzIPN(map[string]string{"status": "VALID"}, "pass", map[string]string{})
	assert.Error(t, err)
}

func TestValidateSSLCommerzIPN_RejectsBadSignature(t *testing.T) {
	params := map[string]string{"status": "VALID", "amount": "100", "currency": "BDT"}
	err := ValidateSSLCommerzIPN(params, "pass", map[string]string{"verify_sign": "bogus", "verify_key": "amount,currency,status"})
	assert.Error(t, err)
}

func TestValidateSSLCommerzIPN_AcceptsValidSignature(t *testing.T) {
	storePass := "teststorepass"
	hashSeq := "amount=100,currency=BDT,status=VALID"
	mac := hmac.New(sha512.New, []byte(storePass))
	mac.Write([]byte(hashSeq))
	expected := hex.EncodeToString(mac.Sum(nil))
	params := map[string]string{"status": "VALID", "amount": "100", "currency": "BDT"}
	verify := map[string]string{"verify_sign": expected, "verify_key": "amount,currency,status"}
	assert.NoError(t, ValidateSSLCommerzIPN(params, storePass, verify))
}