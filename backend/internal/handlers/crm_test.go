package handlers

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/services"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:crm-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.User{}))
	return db
}

func seedUser(t *testing.T, db *gorm.DB, username, email, phone string) domain.User {
	t.Helper()
	user := domain.User{
		Username:    username,
		Email:       email,
		PhoneNumber: phone,
		Password:    "secret-hash",
	}
	require.NoError(t, db.Create(&user).Error)
	return user
}

func TestGetCustomer_Uniform404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	seedUser(t, db, "alice", "a@x.com", "555-1234")

	r := gin.New()
	h := &CRMHandler{db: db}
	r.GET("/crm/customers/:identifier", h.GetCustomer)

	for _, id := range []string{"555-9999", "missing@x.com", "ghost", "9999"} {
		req := httptest.NewRequest("GET", "/crm/customers/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code, "missing identifier %s should be 404", id)
		assert.Contains(t, w.Body.String(), "not found")
	}
}

func TestGetCustomer_ValidLookupsStillWork(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	alice := seedUser(t, db, "alice", "a@x.com", "555-1234")

	r := gin.New()
	h := &CRMHandler{db: db}
	r.GET("/crm/customers/:identifier", h.GetCustomer)

	for _, id := range []string{
		fmt.Sprintf("%d", alice.ID),
		"a@x.com",
		"555-1234",
		"alice",
	} {
		req := httptest.NewRequest("GET", "/crm/customers/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "identifier %s should be 200", id)
		assert.NotContains(t, w.Body.String(), alice.Password)
	}
}

type stubCRMService struct {
	services.CRMService
	risk *domain.ChurnRisk
}

func (s *stubCRMService) GetChurnRisk(customerID uint) (*domain.ChurnRisk, error) {
	return s.risk, nil
}

func newChurnRiskRouter(t *testing.T, db *gorm.DB, mock services.CRMService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &CRMHandler{db: db, crmService: mock}
	r.GET("/crm/customers/:identifier/churn-risk", h.GetChurnRisk)
	return r
}

func TestGetChurnRisk_AcceptsEmail(t *testing.T) {
	db := setupTestDB(t)
	seedUser(t, db, "alice", "a@x.com", "555-1234")
	mock := &stubCRMService{risk: &domain.ChurnRisk{RiskLevel: "low"}}

	r := newChurnRiskRouter(t, db, mock)
	req := httptest.NewRequest("GET", "/crm/customers/a@x.com/churn-risk", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "email identifier should work, got body: %s", w.Body.String())
	assert.Contains(t, w.Body.String(), `"risk_level":"low"`, "should return the churn risk body")
}

func TestGetChurnRisk_AcceptsPhone(t *testing.T) {
	db := setupTestDB(t)
	seedUser(t, db, "alice", "a@x.com", "555-1234")
	mock := &stubCRMService{risk: &domain.ChurnRisk{RiskLevel: "low"}}

	r := newChurnRiskRouter(t, db, mock)
	req := httptest.NewRequest("GET", "/crm/customers/555-1234/churn-risk", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "phone identifier should work, got body: %s", w.Body.String())
	assert.Contains(t, w.Body.String(), `"risk_level":"low"`, "should return the churn risk body")
}

func TestGetChurnRisk_AcceptsUsername(t *testing.T) {
	db := setupTestDB(t)
	seedUser(t, db, "alice", "a@x.com", "555-1234")
	mock := &stubCRMService{risk: &domain.ChurnRisk{RiskLevel: "low"}}

	r := newChurnRiskRouter(t, db, mock)
	req := httptest.NewRequest("GET", "/crm/customers/alice/churn-risk", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "username identifier should work, got body: %s", w.Body.String())
	assert.Contains(t, w.Body.String(), `"risk_level":"low"`, "should return the churn risk body")
}
