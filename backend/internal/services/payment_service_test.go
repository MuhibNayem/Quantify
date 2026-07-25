package services

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSSLCommerzIPN_AcceptsValidSignature(t *testing.T) {
	// Compute the md5 the way the protocol does, then assert the function accepts it.
	storePass := "teststorepass"
	postParams := map[string]string{
		"amount":   "100",
		"currency": "BDT",
		"status":   "VALID",
	}
	// verify_key lists the fields the merchant uses for signing.
	verifyKey := "amount,currency,status"
	// Build the join string with store_passwd=md5(storePass):
	storePasswdHash := fmt.Sprintf("%x", md5.Sum([]byte(storePass)))
	signKeys := map[string]string{
		"amount":       "100",
		"currency":     "BDT",
		"status":       "VALID",
		"store_passwd": storePasswdHash,
	}
	keys := []string{"amount", "currency", "status", "store_passwd"}
	sort.Strings(keys)
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, k+"="+signKeys[k])
	}
	expected := fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(pairs, "&"))))

	verify := map[string]string{"verify_sign": expected, "verify_key": verifyKey}
	assert.NoError(t, ValidateSSLCommerzIPN(postParams, storePass, verify))
}

func TestValidateSSLCommerzIPN_RejectsBadSignature(t *testing.T) {
	postParams := map[string]string{"status": "VALID", "amount": "100"}
	verify := map[string]string{"verify_sign": "deadbeef", "verify_key": "amount,status"}
	err := ValidateSSLCommerzIPN(postParams, "pass", verify)
	assert.Error(t, err)
}

func TestValidateSSLCommerzIPN_RejectsMissingSignature(t *testing.T) {
	err := ValidateSSLCommerzIPN(map[string]string{"status": "VALID"}, "pass", map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verify_sign")
}

func TestValidateSSLCommerzIPN_RejectsInvalidStatus(t *testing.T) {
	// Even with a valid signature, status="FAILED" must be rejected.
	storePass := "test"
	postParams := map[string]string{"status": "FAILED", "amount": "100", "currency": "BDT"}
	storePasswdHash := fmt.Sprintf("%x", md5.Sum([]byte(storePass)))
	signKeys := map[string]string{
		"amount":       "100",
		"currency":     "BDT",
		"status":       "FAILED",
		"store_passwd": storePasswdHash,
	}
	keys := []string{"amount", "currency", "status", "store_passwd"}
	sort.Strings(keys)
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, k+"="+signKeys[k])
	}
	expected := fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(pairs, "&"))))
	err := ValidateSSLCommerzIPN(postParams, storePass, map[string]string{"verify_sign": expected, "verify_key": "amount,currency,status"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status")
}
