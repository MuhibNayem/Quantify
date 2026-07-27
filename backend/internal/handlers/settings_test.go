package handlers

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/services"
)

type stubSettingsService struct {
	mu       sync.Mutex
	store    map[string]string
	updateOK bool
}

func newStubSettingsService() *stubSettingsService {
	return &stubSettingsService{
		store:    make(map[string]string),
		updateOK: true,
	}
}

func (s *stubSettingsService) GetAllSettings() (map[string][]domain.SystemSetting, error) {
	return nil, nil
}
func (s *stubSettingsService) GetSetting(key string) (string, error) {
	return "", nil
}
func (s *stubSettingsService) GetPublicSettings() (map[string]string, error) {
	return nil, nil
}
func (s *stubSettingsService) UpdateSetting(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.updateOK {
		return assert.AnError
	}
	s.store[key] = value
	return nil
}

func (s *stubSettingsService) last(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.store[key]
	return v, ok
}

func putSettingJSON(t *testing.T, h *SettingsHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/settings", h.UpdateSetting)
	req := httptest.NewRequest("PUT", "/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPutSetting_RejectsInvalidTaxRate(t *testing.T) {
	h := &SettingsHandler{service: newStubSettingsService()}
	w := putSettingJSON(t, h, `{"key":"tax_rate_percentage","value":"abc"}`)
	assert.Equal(t, 400, w.Code, "non-numeric tax_rate_percentage should be rejected")
}

func TestPutSetting_RejectsOutOfRangeTaxRate(t *testing.T) {
	h := &SettingsHandler{service: newStubSettingsService()}
	w := putSettingJSON(t, h, `{"key":"tax_rate_percentage","value":"150"}`)
	assert.Equal(t, 400, w.Code, "tax_rate_percentage above 100 must be rejected")
}

func TestPutSetting_AcceptsValidTaxRate(t *testing.T) {
	svc := newStubSettingsService()
	h := &SettingsHandler{service: svc}
	w := putSettingJSON(t, h, `{"key":"tax_rate_percentage","value":"10.5"}`)
	assert.Equal(t, 200, w.Code)
	v, ok := svc.last("tax_rate_percentage")
	assert.True(t, ok, "value should persist on valid input")
	assert.Equal(t, "10.5", v)
}

func TestPutSetting_RejectsUnknownKey(t *testing.T) {
	h := &SettingsHandler{service: newStubSettingsService()}
	w := putSettingJSON(t, h, `{"key":"unknown_setting","value":"x"}`)
	assert.Equal(t, 400, w.Code, "unknown setting key must be rejected")
}

func TestPutSetting_RejectsInvalidCurrency(t *testing.T) {
	h := &SettingsHandler{service: newStubSettingsService()}
	w := putSettingJSON(t, h, `{"key":"currency","value":"USDX"}`)
	assert.Equal(t, 400, w.Code, "currency must be a 3-letter code")
}

func TestPutSetting_AcceptsValidCurrency(t *testing.T) {
	svc := newStubSettingsService()
	h := &SettingsHandler{service: svc}
	w := putSettingJSON(t, h, `{"key":"currency","value":"USD"}`)
	assert.Equal(t, 200, w.Code)
}

func TestPutSetting_RejectsInvalidEarningRate(t *testing.T) {
	h := &SettingsHandler{service: newStubSettingsService()}
	w := putSettingJSON(t, h, `{"key":"loyalty.earning_rate","value":"1.5"}`)
	assert.Equal(t, 400, w.Code, "earning rate must be between 0 and 1")
}

var _ services.SettingsService = (*stubSettingsService)(nil)
