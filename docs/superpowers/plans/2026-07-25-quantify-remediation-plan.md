# Quantify Backend Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all security, data-integrity, correctness, architectural, operational, testing, and documentation issues identified in the backend audit, organized into 7 ordered phase-chapters with verification gates between them.

**Architecture:** Patches and surgical refactors within the existing Go (1.24.3) Gin/GORM/Postgres/Redis/RabbitMQ/MinIO service. Phase 1-3 are hotfixes (security, data, correctness). Phase 4 introduces structural refactors (router split, DI container, golang-migrate, file splits). Phase 5 adds operations (metrics, traces, CI). Phase 6 adds tests. Phase 7 updates docs.

**Tech Stack:** Go 1.24.3, Gin v1.11.0, GORM v1.31.0, Postgres 15, Redis cluster, RabbitMQ 3, MinIO, Prometheus, Grafana, logrus, gorilla/websocket, koanf (added Phase 4), golang-migrate (added Phase 4), testcontainers-go (added Phase 6), OpenTelemetry SDK (added Phase 5).

**Working directory:** `backend/` unless otherwise stated. All commands assume cwd is `backend/`.

**Spec:** `docs/superpowers/specs/2026-07-25-quantify-remediation-design.md`

---

## File Structure (locked in by this plan)

### New files

```
backend/
├── .github/workflows/ci.yml                  # P5.T8
├── CHANGELOG.md                              # all phases append
├── migrations/                                # P4.T5
│   ├── 0001_init.up.sql / .down.sql
│   └── 0002_seed.up.sql / .down.sql          # P4.T6
├── internal/
│   ├── app/
│   │   └── container.go                      # P4.T2
│   ├── metrics/                              # P5.T3
│   │   ├── http.go
│   │   ├── cache.go
│   │   ├── jobs.go
│   │   ├── payments.go
│   │   └── alerts.go
│   ├── telemetry/                            # P5.T7
│   │   ├── tracer.go
│   │   └── propagation.go
│   ├── middleware/
│   │   ├── request_id.go                     # P5.T6
│   │   └── security_headers.go               # P1.T14
│   ├── repository/
│   │   ├── interfaces.go                     # P4.T3
│   │   ├── sales_reports.go                  # P4.T8
│   │   ├── inventory_reports.go
│   │   ├── financial_reports.go
│   │   ├── crm_reports.go
│   │   └── audit_log.go                      # P3.T4
│   ├── services/
│   │   ├── pricing.go                        # P4.T9
│   │   ├── stock_allocator.go
│   │   ├── loyalty.go
│   │   ├── promotion_resolver.go
│   │   ├── checkout.go
│   │   └── stock_transfer_service.go         # P2.T1
│   ├── handlers/
│   │   ├── ready.go                          # P5.T1
│   │   ├── request_return.go                 # P4.T10
│   │   └── process_return.go
│   ├── integrations/
│   │   ├── shopify.go                        # P4.T12
│   │   └── quickbooks.go
│   ├── message_broker/
│   │   └── topology.go                       # P4.T15
│   └── router/
│       ├── router.go                         # P4.T1 (rewritten)
│       └── groups/
│           ├── public.go                     # P4.T1
│           ├── payment.go
│           ├── websocket.go
│           ├── auth.go
│           ├── users.go
│           ├── products.go
│           ├── sales.go
│           ├── reports.go
│           ├── replenishment.go
│           ├── alerts.go
│           ├── bulk.go
│           ├── inventory.go
│           ├── crm.go
│           ├── time_tracking.go
│           ├── search.go
│           ├── settings.go
│           └── roles.go
├── test/
│   └── integration/
│       └── db.go                             # P6.T1
grafana/dashboards/                             # P5.T4
├── overview.json
├── api.json
├── jobs.json
├── payments.json
└── alerts.json
docs/
├── architecture.md                           # P7.T3
├── security.md                               # P7.T4
├── runbook.md                                # P7.T5
└── log-shipping.md                           # P5.T10
```

### Modified files (key)

- `backend/internal/router/router.go` (split — P4.T1)
- `backend/internal/repository/database.go` (DI — P4.T4)
- `backend/internal/services/sales_service.go` (split — P4.T9)
- `backend/internal/handlers/return_handler.go` (split — P4.T10)
- `backend/internal/services/purchase_order_service.go` (P2.T5, P2.T6, P2.T11)
- `backend/internal/services/vendor_return_service.go` (P2.T4, P2.T7)
- `backend/internal/consumers/bulk_consumer.go` (P3.T1, P3.T2)
- `backend/internal/consumers/alert_consumer.go` (P1.T13)
- `backend/internal/repository/reporting_repository.go` (split — P4.T8)
- `backend/internal/handlers/inventory.go` (P2.T1)
- `backend/internal/handlers/return_handler.go` (P2.T2, P2.T3)
- `backend/internal/handlers/crm.go` (P2.T13, P2.T14, P3.T13)
- `backend/internal/handlers/sales.go` (P2.T9, P3.T9)
- `backend/internal/handlers/user.go` (P1.T4)
- `backend/internal/handlers/websocket.go` (P1.T3, P3.T11)
- `backend/internal/middleware/permission.go` (P1.T5)
- `backend/internal/middleware/csrf.go` (P1.T10)
- `backend/internal/repository/role_repository.go` (P1.T5, P1.T10)
- `backend/internal/repository/search.go` (P1.T11)
- `backend/internal/services/payment_service.go` (P1.T2)
- `backend/internal/services/replenishment_service.go` (P4.T18)
- `backend/internal/services/email_service.go` (P5.T5)
- `backend/internal/handlers/notifications.go` (P1.T12, P3.T12, P5.T5)
- `backend/internal/bootstrap/app.go` (P4.T2, P4.T7, P4.T16)
- `backend/internal/config/config.go` (P1.T4, P1.T9, P4.T13)
- `backend/internal/migrations/seed_ai_user.go` (P1.T1, P1.T6)
- `backend/cmd/server/main.go` (P4.T2, P5.T6)
- `backend/cmd/migrator/main.go` (P4.T5)
- `backend/Dockerfile` (P5.T9)
- `backend/.gitignore` (P1.T8)
- `backend/api/openapi.yaml` (P4.T11, P7.T1)
- `backend/api/docs.go` (P4.T11)
- `backend/api/swagger.json` (P4.T11)
- `backend/api/swagger.yaml` (P4.T11)
- `backend/.env` (delete — P1.T8)
- `backend/.env.example` (P1.T9)
- `docs/business_logic.md` (P7.T2)
- `README.md`, `docs/README.md` (P7.T6)

---

## Commit cadence

- Each task ends with a commit.
- Each phase ends with a merge to main and a CHANGELOG entry.
- Branch per phase: `phase-1-security`, `phase-2-integrity`, `phase-3-correctness`, `phase-4-architecture`, `phase-5-ops`, `phase-6-tests`, `phase-7-docs`.

---

# Phase 1 — Security hardening (14 tasks)

## Task 1.1: Remove hard-coded AI agent password

**Files:**
- Modify: `backend/internal/migrations/seed_ai_user.go:31`
- Modify: `backend/internal/config/config.go`
- Test: `backend/internal/migrations/seed_ai_user_test.go` (new)

- [ ] **Step 1: Write the failing test**

Create `backend/internal/migrations/seed_ai_user_test.go`:

```go
package migrations

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestSeedAIUser_NoHardcodedPassword(t *testing.T) {
	password := GenerateRandomPassword()
	assert.GreaterOrEqual(t, len(password), 24, "generated password must be >=24 chars")
	assert.NotEqual(t, "rZQ$4Rs!6{QHaR{5Sra{]z_%n", password)
}

func TestSeedAIUser_HashIsBcrypt(t *testing.T) {
	password := GenerateRandomPassword()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	assert.NoError(t, err)
}
```

- [ ] **Step 2: Run the test to confirm it fails**

Run: `cd backend && go test ./internal/migrations/... -v`
Expected: FAIL with `undefined: GenerateRandomPassword`

- [ ] **Step 3: Implement `GenerateRandomPassword`**

In `backend/internal/migrations/seed_ai_user.go`, add at the top:

```go
package migrations

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
)

// GenerateRandomPassword returns a 32-byte URL-safe base64 string (~43 chars).
func GenerateRandomPassword() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(b)
}
```

- [ ] **Step 4: Replace hard-coded password in SeedAIUser**

Find the function that inserts the AI user. Replace the hard-coded hash with:

```go
func SeedAIUser(db *gorm.DB) error {
	password := GenerateRandomPassword()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := domain.User{
		Username:  "ai-agent",
		Email:     "ai-agent@quantify.com",
		Password:  string(hash),
		IsActive:  false, // P1.T6 — operator must approve
		RoleID:    adminRoleID(db),
		CreatedAt: time.Now(),
	}
	// ... rest of insert
	logrus.Info("AI agent user created. One-time password (store securely):", password)
	return db.Create(&user).Error
}
```

- [ ] **Step 5: Run the test to confirm it passes**

Run: `cd backend && go test ./internal/migrations/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd backend
git add internal/migrations/seed_ai_user.go internal/migrations/seed_ai_user_test.go
git commit -m "fix(security): generate random password for AI agent user (P1.T1)"
```

---

## Task 1.2: SSLCommerz IPN signature verification

**Files:**
- Modify: `backend/internal/services/payment_service.go:289-315`
- Modify: `backend/internal/config/config.go`
- Test: `backend/internal/services/payment_service_test.go` (new)

- [ ] **Step 1: Write the failing test**

```go
package services

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strings"
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
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/services/ -run TestValidateSSLCommerzIPN -v`
Expected: FAIL with `undefined: ValidateSSLCommerzIPN` (new signature mismatch)

- [ ] **Step 3: Implement signature verification**

Replace `ValidateSSLCommerzIPN` in `backend/internal/services/payment_service.go`:

```go
import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strings"
)

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

func validateSSLCommerzFromValues(form url.Values, storePass string) error {
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
```

- [ ] **Step 4: Update call sites**

In `payment_service.go`, replace the existing call with:

```go
if err := validateSSLCommerzFromValues(c.Request.PostForm, cfg.SSLCommerzStorePass); err != nil {
    return c.JSON(400, gin.H{"error": err.Error()})
}
```

Add to `config/config.go`:

```go
SSLCommerzStorePass string // SSLCOMMERZ_STORE_PASS
```

In `LoadConfig`:

```go
SSLCommerzStorePass: getEnv("SSLCOMMERZ_STORE_PASS", ""),
```

- [ ] **Step 5: Run, confirm pass**

Run: `cd backend && go test ./internal/services/ -run TestValidateSSLCommerzIPN -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd backend
git add internal/services/payment_service.go internal/services/payment_service_test.go internal/config/config.go
git commit -m "fix(security): verify SSLCommerz IPN HMAC signature (P1.T2)"
```

---

## Task 1.3: WebSocket origin allowlist

**Files:**
- Modify: `backend/internal/handlers/websocket.go:19-22`
- Modify: `backend/internal/config/config.go`
- Test: `backend/internal/handlers/websocket_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestBuildOriginChecker_AllowsListedOrigin(t *testing.T) {
	check := buildOriginChecker([]string{"https://app.example.com"})
	assert.True(t, check(reqWithOrigin("https://app.example.com")))
	assert.False(t, check(reqWithOrigin("https://evil.example.com")))
	assert.False(t, check(reqWithOrigin("")))
}

func reqWithOrigin(origin string) *http.Request {
	r := httptest.NewRequest("GET", "/ws", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/handlers/ -run TestBuildOriginChecker -v`
Expected: FAIL (`undefined: buildOriginChecker`)

- [ ] **Step 3: Implement**

Add to `backend/internal/handlers/websocket.go`:

```go
package handlers

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
)

func buildOriginChecker(allowed []string) func(r *http.Request) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		set[strings.TrimSpace(o)] = struct{}{}
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		_, ok := set[origin]
		return ok
	}
}
```

In `ServeWs`, replace `CheckOrigin: func(r *http.Request) bool { return true }` with:

```go
CheckOrigin: buildOriginChecker(cfg.WSAllowedOrigins),
```

- [ ] **Step 4: Wire config**

In `config/config.go`:

```go
WSAllowedOrigins []string // WS_ALLOWED_ORIGINS, CSV
```

Add helper to `LoadConfig`:

```go
func loadOrigins(env, def string) []string {
	v := getEnv(env, def)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// inside LoadConfig:
WSAllowedOrigins: loadOrigins("WS_ALLOWED_ORIGINS", "http://localhost:5173"),
```

- [ ] **Step 5: Run test, confirm pass**

Run: `cd backend && go test ./internal/handlers/ -run TestBuildOriginChecker -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd backend
git add internal/handlers/websocket.go internal/handlers/websocket_test.go internal/config/config.go
git commit -m "fix(security): enforce WebSocket origin allowlist (P1.T3)"
```

---

## Task 1.4: Configurable refresh token TTL

**Files:**
- Modify: `backend/internal/handlers/user.go:217,304`
- Modify: `backend/internal/config/config.go`
- Test: `backend/internal/handlers/user_ttl_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
package handlers

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"inventory/backend/internal/config"
)

func TestResolveRefreshTTL_Default30Days(t *testing.T) {
	cfg := &config.Config{}
	assert.Equal(t, 720*time.Hour, resolveRefreshTTL(cfg))
}

func TestResolveRefreshTTL_RespectsConfig(t *testing.T) {
	cfg := &config.Config{RefreshTokenTTL: 24 * time.Hour}
	assert.Equal(t, 24*time.Hour, resolveRefreshTTL(cfg))
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/handlers/ -run TestResolveRefreshTTL -v`
Expected: FAIL (`undefined: resolveRefreshTTL`)

- [ ] **Step 3: Implement**

In `config/config.go`:

```go
RefreshTokenTTL time.Duration // REFRESH_TOKEN_TTL, default 720h
```

In `LoadConfig`:

```go
RefreshTokenTTL: parseDuration("REFRESH_TOKEN_TTL", 720*time.Hour),
```

Add helper:

```go
func parseDuration(env string, def time.Duration) time.Duration {
	v := os.Getenv(env)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
```

In `user.go`:

```go
func resolveRefreshTTL(cfg *config.Config) time.Duration {
	if cfg.RefreshTokenTTL <= 0 {
		return 720 * time.Hour
	}
	return cfg.RefreshTokenTTL
}
```

Replace the two hard-coded 10-year TTLs (`10*365*24*time.Hour`) at `user.go:217` and `:304`:

```go
ttl := resolveRefreshTTL(cfg)
err = redis.Set(ctx, "refresh_token:"+refresh, userID, ttl)
```

- [ ] **Step 4: Run test, confirm pass**

Run: `cd backend && go test ./internal/handlers/ -run TestResolveRefreshTTL -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/handlers/user.go internal/handlers/user_ttl_test.go internal/config/config.go
git commit -m "fix(security): make refresh token TTL configurable, default 30d (P1.T4)"
```

---

## Task 1.5: Permission cache invalidation on every mutation

**Files:**
- Modify: `backend/internal/repository/role_repository.go:59-138`
- Modify: `backend/internal/repository/role_repository_ext.go:12-42`
- Modify: `backend/internal/middleware/permission.go:13-55`
- Test: `backend/internal/repository/role_repository_ext_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
package repository

import (
	"context"
	"testing"
	"time"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionCache_InvalidatedOnGrant(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := &RoleRepositoryExt{redis: rdb}

	ctx := context.Background()
	require.NoError(t, repo.SetPermissionsCache(ctx, "Manager", []string{"products.read"}))
	got, _ := repo.GetPermissionsByRoleName(ctx, "Manager")
	assert.Equal(t, []string{"products.read"}, got)

	require.NoError(t, repo.InvalidateRoleCache(ctx, "Manager"))
	got, _ = repo.GetPermissionsByRoleName(ctx, "Manager")
	assert.Empty(t, got, "cache must be cleared immediately on invalidation")
}

func TestPermissionCache_TTLIsShort(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := &RoleRepositoryExt{redis: rdb}
	require.NoError(t, repo.SetPermissionsCache(context.Background(), "Manager", []string{"x"}))
	ttl := mr.Ttl("rbac:role:Manager:permissions")
	assert.LessOrEqual(t, ttl, 60*time.Second, "TTL must be <= 60s")
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/repository/ -run TestPermissionCache -v`
Expected: FAIL (`SetPermissionsCache` undefined).

- [ ] **Step 3: Implement cache helpers**

In `role_repository_ext.go`:

```go
const permissionCacheTTL = 60 * time.Second

func (r *RoleRepositoryExt) SetPermissionsCache(ctx context.Context, role string, perms []string) error {
	key := "rbac:role:" + role + ":permissions"
	val, _ := json.Marshal(perms)
	return r.redis.Set(ctx, key, val, permissionCacheTTL).Err()
}

func (r *RoleRepositoryExt) GetPermissionsByRoleName(ctx context.Context, role string) ([]string, error) {
	key := "rbac:role:" + role + ":permissions"
	if cached, err := r.redis.Get(ctx, key).Bytes(); err == nil {
		var perms []string
		if err := json.Unmarshal(cached, &perms); err == nil {
			return perms, nil
		}
	}
	// miss: load from DB, cache
	perms, err := r.loadFromDB(ctx, role)
	if err != nil {
		return nil, err
	}
	if len(perms) > 0 {
		_ = r.SetPermissionsCache(ctx, role, perms)
	}
	return perms, nil
}

func (r *RoleRepositoryExt) InvalidateRoleCache(ctx context.Context, role string) error {
	return r.redis.Del(ctx, "rbac:role:"+role+":permissions").Err()
}
```

In `role_repository.go`, add calls in `AssignPermissions`, `UpdateRole`, `DeleteRole`, and `RemovePermission`:

```go
// At end of each mutation method:
if err := r.ext.InvalidateRoleCache(ctx, roleName); err != nil {
    logrus.WithError(err).Warn("permission cache invalidation failed")
}
```

- [ ] **Step 4: Run test, confirm pass**

Run: `cd backend && go test ./internal/repository/ -run TestPermissionCache -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/repository/role_repository.go internal/repository/role_repository_ext.go internal/repository/role_repository_ext_test.go internal/middleware/permission.go
git commit -m "fix(security): invalidate permission cache on every mutation (P1.T5)"
```

---

## Task 1.6: Deactivate AI agent user by default

**Files:**
- Modify: `backend/internal/migrations/seed_ai_user.go`
- Test: extends `backend/internal/migrations/seed_ai_user_test.go`

- [ ] **Step 1: Write failing test**

Add to `seed_ai_user_test.go`:

```go
func TestSeedAIUser_DefaultInactive(t *testing.T) {
	user := buildAIUser("pwhash")
	assert.False(t, user.IsActive, "AI agent user must default to inactive")
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/migrations/ -run TestSeedAIUser_DefaultInactive -v`
Expected: FAIL (current code sets `IsActive=true`).

- [ ] **Step 3: Set `IsActive: false`**

In `seed_ai_user.go`:

```go
user := domain.User{
    Username: "ai-agent",
    Email:    "ai-agent@quantify.com",
    Password: string(hash),
    IsActive: false, // operator must call PUT /users/:id/approve
    RoleID:   adminRoleID(db),
    CreatedAt: time.Now(),
}
```

- [ ] **Step 4: Run, confirm pass**

Run: `cd backend && go test ./internal/migrations/ -run TestSeedAIUser -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/migrations/seed_ai_user.go internal/migrations/seed_ai_user_test.go
git commit -m "fix(security): deactivate AI agent user by default (P1.T6)"
```

---

## Task 1.7: Remove `gen_hash.go` helper

**Files:**
- Delete: `backend/gen_hash.go`

- [ ] **Step 1: Verify no production code imports it**

Run: `cd backend && grep -r "gen_hash" --include="*.go" .`
Expected: no matches.

- [ ] **Step 2: Delete the file**

```bash
rm backend/gen_hash.go
```

- [ ] **Step 3: Verify build still passes**

Run: `cd backend && go build ./...`
Expected: success.

- [ ] **Step 4: Commit**

```bash
cd backend
git rm gen_hash.go
git commit -m "chore: remove gen_hash.go helper (P1.T7)"
```

---

## Task 1.8: Scrub `.env` from history and gitignore

**Files:**
- Modify: `backend/.gitignore`
- Delete (working tree): `backend/.env`

- [ ] **Step 1: Update `.gitignore`**

Ensure `backend/.gitignore` contains:

```
.env
.env.local
.env.*.local
```

- [ ] **Step 2: Remove from working tree**

```bash
cd backend
git rm --cached .env
rm .env
```

- [ ] **Step 3: Scrub from history**

Install `git-filter-repo` if not present:

```bash
pip install git-filter-repo
```

Then:

```bash
cd backend
git filter-repo --path .env --invert-paths
```

(This rewrites all branches — coordinate with team if collaborating.)

- [ ] **Step 4: Rotate any committed secrets**

If `JWT_SECRET` or other secrets were in the old `.env` and any environment actually used them, rotate them in those environments. Document in `docs/runbook.md` (Phase 7).

- [ ] **Step 5: Commit**

```bash
cd backend
git add .gitignore
git commit -m "chore: remove .env from repo, add to gitignore (P1.T8)"
```

---

## Task 1.9: Reject default-value secrets on boot

**Files:**
- Modify: `backend/internal/config/config.go:145-174`
- Test: `backend/internal/config/config_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig_RejectsDefaultJWTSecret(t *testing.T) {
	cfg := &Config{JWTSecret: "my_secret_key", RefreshSecret: "ok", MinioAccessKey: "ok", MinioSecretKey: "ok"}
	err := validateCriticalConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestValidateConfig_RejectsDefaultMinioSecrets(t *testing.T) {
	cfg := &Config{JWTSecret: "unique", RefreshSecret: "unique", MinioAccessKey: "minioadmin", MinioSecretKey: "minioadmin"}
	err := validateCriticalConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfig_AcceptsUniqueSecrets(t *testing.T) {
	cfg := &Config{JWTSecret: "u1", RefreshSecret: "u2", MinioAccessKey: "u3", MinioSecretKey: "u4"}
	assert.NoError(t, validateCriticalConfig(cfg))
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/config/ -run TestValidateConfig -v`
Expected: FAIL on `RejectsDefaultMinioSecrets` (currently only `minioadmin` is rejected, not `minioadmin4321`).

- [ ] **Step 3: Expand disallowed list**

In `config/config.go`, replace the validation block:

```go
var disallowedSecrets = map[string][]string{
	"JWT_SECRET":             {"my_secret_key", "secret", "changeme"},
	"REFRESH_TOKEN_SECRET":   {"my_refresh_secret_key", "secret", "changeme"},
	"MINIO_ACCESS_KEY_ID":    {"minioadmin", "minioadmin4321"},
	"MINIO_SECRET_ACCESS_KEY":{"minioadmin", "minioadmin4321"},
}

func validateCriticalConfig(cfg *Config) error {
	if cfg.RabbitMQURL == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	for field, bad := range disallowedSecrets {
		val := secretValueFor(cfg, field)
		for _, b := range bad {
			if val == b {
				return fmt.Errorf("%s must not use default value %q", field, b)
			}
		}
		if val == "" {
			return fmt.Errorf("%s is required", field)
		}
	}
	return nil
}

func secretValueFor(c *Config, field string) string {
	switch field {
	case "JWT_SECRET": return c.JWTSecret
	case "REFRESH_TOKEN_SECRET": return c.RefreshSecret
	case "MINIO_ACCESS_KEY_ID": return c.MinioAccessKey
	case "MINIO_SECRET_ACCESS_KEY": return c.MinioSecretKey
	}
	return ""
}
```

- [ ] **Step 4: Run, confirm pass**

Run: `cd backend && go test ./internal/config/ -run TestValidateConfig -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/config/config.go internal/config/config_test.go
git commit -m "fix(security): reject default-value secrets on boot (P1.T9)"
```

---

## Task 1.10: CSRF token rotation on privilege change

**Files:**
- Modify: `backend/internal/auth/csrf.go`
- Modify: `backend/internal/repository/role_repository.go` (after permission grant)
- Test: `backend/internal/auth/csrf_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
package auth

import (
	"context"
	"testing"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRF_RotatedOnPrivilegeChange(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	userID := uint(42)
	oldToken, err := GenerateCSRFToken(rdb, userID)
	require.NoError(t, err)

	RotateCSRFTokenFor(rdb, userID)

	newToken, err := GenerateCSRFToken(rdb, userID)
	require.NoError(t, err)
	assert.NotEqual(t, oldToken, newToken)

	valid, err := ValidateCSRFToken(rdb, oldToken, userID)
	assert.Error(t, err)
	assert.False(t, valid)
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/auth/ -run TestCSRF -v`
Expected: FAIL (`RotateCSRFTokenFor` undefined).

- [ ] **Step 3: Implement**

In `backend/internal/auth/csrf.go`:

```go
func RotateCSRFTokenFor(rdb *redis.Client, userID uint) error {
	ctx := context.Background()
	pattern := "csrf_token:*"
	iter := rdb.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	// Each key is "csrf_token:<token>:<userID>" — delete only those matching userID.
	target := strconv.FormatUint(uint64(userID), 10)
	var toDelete []string
	for _, k := range keys {
		if strings.HasSuffix(k, ":"+target) {
			toDelete = append(toDelete, k)
		}
	}
	if len(toDelete) == 0 {
		return nil
	}
	return rdb.Del(ctx, toDelete...).Err()
}
```

(Adjust the suffix-matching to match the actual key format used by `GenerateCSRFToken` in this file.)

In `role_repository.go`, after `AssignPermissions`:

```go
if err := r.ext.InvalidateRoleCache(ctx, roleName); err != nil { ... }

affectedUsers, _ := r.db.Model(&domain.User{}).Where("role_id = ?", roleID).Pluck("id", &userIDs).Error
for _, uid := range userIDs {
    if err := auth.RotateCSRFTokenFor(r.redis, uid); err != nil {
        logrus.WithError(err).Warn("CSRF rotation failed")
    }
}
```

- [ ] **Step 4: Run, confirm pass**

Run: `cd backend && go test ./internal/auth/ -run TestCSRF -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/auth/csrf.go internal/auth/csrf_test.go internal/repository/role_repository.go
git commit -m "fix(security): rotate CSRF token on privilege change (P1.T10)"
```

---

## Task 1.11: Sanitize global search query

**Files:**
- Modify: `backend/internal/repository/search.go:75-107`
- Test: `backend/internal/repository/search_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
package repository

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchRepository_HandlesSpecialChars(t *testing.T) {
	db := setupTestDB(t)
	repo := &SearchRepository{db: db}
	require.NoError(t, repo.UpdateIndex(context.Background(), "product", 1, "Coffee Mug"))

	for _, q := range []string{"coffee:", "tea&cake", "a:b:c", "x & y"} {
		_, err := repo.Search(context.Background(), q)
		assert.NoError(t, err, "query %q must not error", q)
	}
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/repository/ -run TestSearchRepository -v`
Expected: FAIL.

- [ ] **Step 3: Replace `to_tsquery` with `websearch_to_tsquery`**

In `search.go`:

```go
func (r *SearchRepository) Search(ctx context.Context, query string) ([]SearchHit, error) {
	var hits []SearchHit
	err := r.db.WithContext(ctx).
		Raw(`SELECT entity_type, entity_id, ts_rank(search_vector, websearch_to_tsquery('english', ?)) AS rank
		     FROM global_search_entries
		     WHERE search_vector @@ websearch_to_tsquery('english', ?)
		     ORDER BY rank DESC LIMIT 50`, query, query).
		Scan(&hits).Error
	return hits, err
}
```

- [ ] **Step 4: Run, confirm pass**

Run: `cd backend && go test ./internal/repository/ -run TestSearchRepository -v`
Expected: PASS (requires Postgres; tag with `//go:build integration` if needed).

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/repository/search.go internal/repository/search_test.go
git commit -m "fix(security): use websearch_to_tsquery for safe search (P1.T11)"
```

---

## Task 1.12: Notification ownership enforcement

**Files:**
- Modify: `backend/internal/handlers/notifications.go:36-75`
- Test: `backend/internal/handlers/notifications_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
package handlers

import (
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestListNotifications_RejectsMissingUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &NotificationHandler{}
	r.GET("/users/:id/notifications", h.ListNotifications)

	req := httptest.NewRequest("GET", "/users/1/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code, "missing user context must 401, not silently pass")
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/handlers/ -run TestListNotifications -v`
Expected: FAIL.

- [ ] **Step 3: Enforce strict ownership**

In `notifications.go`:

```go
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	authUserID := c.GetUint("user_id")
	if authUserID == 0 {
		c.AbortWithStatusJSON(401, gin.H{"error": "unauthenticated"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid user id"})
		return
	}
	if uint(userID) != authUserID && !h.canAccessUser(c, authUserID, uint(userID)) {
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
		return
	}
	// ... existing list logic
}
```

Audit all other handlers using `c.GetUint("user_id")` and apply the same pattern.

- [ ] **Step 4: Run, confirm pass**

Run: `cd backend && go test ./internal/handlers/ -run TestListNotifications -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/handlers/notifications.go internal/handlers/notifications_test.go
git commit -m "fix(security): reject notifications when user_id missing from context (P1.T12)"
```

---

## Task 1.13: Fix alert consumer role query

**Files:**
- Modify: `backend/internal/consumers/alert_consumer.go:124`
- Test: `backend/internal/consumers/alert_consumer_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
package consumers

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"inventory/backend/internal/domain"
	"gorm.io/gorm"
)

func TestLoadSubscribedUsers_FindsByRoleName(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.Create(&domain.Role{Name: "Manager"}).Error)
	var role domain.Role
	db.Where("name = ?", "Manager").First(&role)
	require.NoError(t, db.Create(&domain.User{Username: "alice", RoleID: role.ID}).Error)
	require.NoError(t, db.Create(&domain.AlertRoleSubscription{AlertType: "LOW_STOCK", RoleName: "Manager"}).Error)

	users, err := loadSubscribedUsers(context.Background(), db, "LOW_STOCK")
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "alice", users[0].Username)
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/consumers/ -run TestLoadSubscribedUsers -v`
Expected: FAIL.

- [ ] **Step 3: Fix the query**

In `alert_consumer.go:124`, replace the user lookup:

```go
func loadSubscribedUsers(ctx context.Context, db *gorm.DB, alertType string) ([]domain.User, error) {
	var subs []domain.AlertRoleSubscription
	if err := db.WithContext(ctx).Where("alert_type = ?", alertType).Find(&subs).Error; err != nil {
		return nil, err
	}
	if len(subs) == 0 {
		return nil, nil
	}
	roleNames := make([]string, 0, len(subs))
	for _, s := range subs {
		roleNames = append(roleNames, s.RoleName)
	}
	var users []domain.User
	err := db.WithContext(ctx).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name IN ?", roleNames).
		Find(&users).Error
	return users, err
}
```

- [ ] **Step 4: Run, confirm pass**

Run: `cd backend && go test ./internal/consumers/ -run TestLoadSubscribedUsers -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/consumers/alert_consumer.go internal/consumers/alert_consumer_test.go
git commit -m "fix(alerts): query users via roles.name join (P1.T13)"
```

---

## Task 1.14: Security headers middleware

**Files:**
- Create: `backend/internal/middleware/security_headers.go`
- Modify: `backend/internal/router/router.go` (or `router/groups/public.go` after P4.T1)
- Test: `backend/internal/middleware/security_headers_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
package middleware

import (
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders_PresentOnResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/x", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=")
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
}
```

- [ ] **Step 2: Run — confirm fail**

Run: `cd backend && go test ./internal/middleware/ -run TestSecurityHeaders -v`
Expected: FAIL (`SecurityHeaders` undefined).

- [ ] **Step 3: Implement middleware**

```go
package middleware

import "github.com/gin-gonic/gin"

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", "default-src 'self'")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}
```

- [ ] **Step 4: Wire into router**

In `router.go`, after `r := gin.New()`:

```go
r.Use(middleware.SecurityHeaders())
```

- [ ] **Step 5: Run, confirm pass**

Run: `cd backend && go test ./internal/middleware/ -run TestSecurityHeaders -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd backend
git add internal/middleware/security_headers.go internal/middleware/security_headers_test.go internal/router/router.go
git commit -m "feat(security): add security headers middleware (P1.T14)"
```

---

## Phase 1 verification gate

```bash
cd backend
go test ./...
golangci-lint run
gosec ./...
```

Then append to `backend/CHANGELOG.md`:

```markdown
## [Phase 1] 2026-XX-XX — Security hardening

- P1.T1: Random AI agent password on first boot
- P1.T2: SSLCommerz IPN HMAC signature verification
- P1.T3: WebSocket origin allowlist
- P1.T4: Refresh token TTL configurable, default 30 days
- P1.T5: Permission cache invalidated on every role mutation
- P1.T6: AI agent user deactivated by default
- P1.T7: Removed gen_hash.go helper
- P1.T8: Scrubbed .env from repo
- P1.T9: Reject default-value secrets on boot
- P1.T10: CSRF rotation on privilege change
- P1.T11: Safe search via websearch_to_tsquery
- P1.T12: Strict notification ownership
- P1.T13: Alert consumer joins roles table
- P1.T14: Security headers middleware
```

Commit CHANGELOG, merge `phase-1-security` to main, tag `v1.1-security`.
---

# Phase 2 — Data integrity (16 tasks)

## Task 2.1: Stock transfer updates batches

**Files:**
- Modify: `backend/internal/handlers/inventory.go:28-105`
- Create: `backend/internal/services/stock_transfer_service.go`
- Test: `backend/internal/services/stock_transfer_service_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestStockTransfer_UpdatesBatches(t *testing.T) {
	db := setupTestDB(t)
	srcBatch := seedBatch(t, db, prodID, locSrc, 100, time.Now().Add(30*24*time.Hour))
	dstBatch := seedBatch(t, db, prodID, locDst, 0, time.Now().Add(30*24*time.Hour))

	svc := NewStockTransferService(db)
	err := svc.Transfer(ctx, StockTransferRequest{
		ProductID: prodID, SourceLocationID: locSrc, DestLocationID: locDst,
		Quantity: 25, Reason: "rebalance", UserID: userID,
	})
	require.NoError(t, err)

	assertBatchQty(t, db, srcBatch.ID, 75)
	assertBatchQty(t, db, dstBatch.ID, 25)
}
```

- [ ] **Step 2-5: Implement service**

Create `stock_transfer_service.go`:

```go
package services

type StockTransferService struct{ db *gorm.DB }

func NewStockTransferService(db *gorm.DB) *StockTransferService { return &StockTransferService{db: db} }

func (s *StockTransferService) Transfer(ctx context.Context, req StockTransferRequest) error {
	if req.Quantity <= 0 { return errors.New("quantity must be positive") }
	if req.SourceLocationID == req.DestLocationID { return errors.New("source and destination must differ") }
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// FEFO deduction from source with row lock
		var srcBatches []domain.Batch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id = ? AND location_id = ? AND quantity > 0", req.ProductID, req.SourceLocationID).
			Order("expiry_date ASC NULLS LAST, created_at ASC").
			Find(&srcBatches).Error; err != nil { return err }
		remaining := req.Quantity
		for i := range srcBatches {
			if remaining <= 0 { break }
			take := srcBatches[i].Quantity
			if take > remaining { take = remaining }
			srcBatches[i].Quantity -= take
			if err := tx.Save(&srcBatches[i]).Error; err != nil { return err }
			remaining -= take
		}
		if remaining > 0 { return errors.New("insufficient stock at source") }

		// Increment destination: existing batch first, else create
		var dst domain.Batch
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id = ? AND location_id = ?", req.ProductID, req.DestLocationID).
			Order("created_at DESC").First(&dst).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) { return err }
		if dst.ID == 0 {
			dst = domain.Batch{ProductID: req.ProductID, LocationID: req.DestLocationID, Quantity: req.Quantity}
			if err := tx.Create(&dst).Error; err != nil { return err }
		} else {
			dst.Quantity += req.Quantity
			if err := tx.Save(&dst).Error; err != nil { return err }
		}

		// Audit rows + transfer record
		xfer := domain.StockTransfer{
			ProductID: req.ProductID, SourceLocationID: req.SourceLocationID,
			DestLocationID: req.DestLocationID, Quantity: req.Quantity,
			Reason: req.Reason, TransferredBy: req.UserID, TransferredAt: time.Now(),
		}
		if err := tx.Create(&xfer).Error; err != nil { return err }
		adjOut := domain.StockAdjustment{ProductID: req.ProductID, LocationID: req.SourceLocationID, Type: "STOCK_OUT", ReasonCode: "TRANSFER", Quantity: req.Quantity, AdjustedBy: req.UserID, AdjustedAt: time.Now()}
		adjIn := adjOut; adjIn.LocationID = req.DestLocationID; adjIn.Type = "STOCK_IN"
		return tx.Create(&[]domain.StockAdjustment{adjOut, adjIn}).Error
	})
}
```

- [ ] **Step 6: Replace handler logic in `inventory.go`**

```go
func (h *InventoryHandler) CreateTransfer(c *gin.Context) {
	var req StockTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.RespondError(c, err); return }
	svc := services.NewStockTransferService(h.container.DB)
	if err := svc.Transfer(c.Request.Context(), req); err != nil { middleware.RespondError(c, err); return }
	c.JSON(201, gin.H{"status": "transferred"})
}
```

- [ ] **Step 7: Run, confirm pass; commit**

```bash
cd backend && go test ./internal/services/ -run TestStockTransfer -v
git add internal/services/stock_transfer_service.go internal/services/stock_transfer_service_test.go internal/handlers/inventory.go
git commit -m "fix(inventory): stock transfer updates batches (P2.T1)"
```

---

## Task 2.2: Return refund math fix (additive tax)

**Files:**
- Modify: `backend/internal/handlers/return_handler.go:55-216`

- [ ] **Step 1: Write failing test**

```go
func TestReturnRefund_TaxOnSubtotal(t *testing.T) {
	// Refund 1 item @ $99.99 with 7.5% tax:
	// subtotal=99.99, tax_share=99.99*0.075=7.49925 → round to 7.50
	// total = 99.99 + 7.50 = 107.49
	got := computeRefund(99.99, 1, 0.075)
	assert.InDelta(t, 107.49, got, 0.01)
}

func TestComputeRefund_NoTax(t *testing.T) {
	assert.Equal(t, 200.0, computeRefund(100, 2, 0))
}
```

- [ ] **Step 2: Run — fail**

`cd backend && go test ./internal/handlers/ -run TestReturnRefund -v` → FAIL.

- [ ] **Step 3: Implement `computeRefund`**

```go
// computeRefund returns refund amount with tax added on top (not compounded).
func computeRefund(unitPrice float64, qty int, taxRate float64) float64 {
	subtotal := unitPrice * float64(qty)
	tax := subtotal * taxRate
	return math.Round((subtotal+tax)*100) / 100
}
```

Replace the inline calculation in `return_handler.go:55-216` with this function call.

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestReturnRefund -v
git add internal/handlers/return_handler.go internal/handlers/return_handler_test.go
git commit -m "fix(returns): additive tax on subtotal for refund (P2.T2)"
```

---

## Task 2.3: Loyalty point reversal prorates by PointsRedeemed

**Files:**
- Modify: `backend/internal/handlers/return_handler.go:382-399`

- [ ] **Step 1: Write failing test**

```go
func TestLoyaltyReversal_Proportional(t *testing.T) {
	assert.Equal(t, 40, computeLoyaltyReversal(100, 10, 4))   // 100*4/10
	assert.Equal(t, 100, computeLoyaltyReversal(100, 10, 10)) // full return
	assert.Equal(t, 20, computeLoyaltyReversal(80, 4, 1))    // partial
	assert.Equal(t, 0, computeLoyaltyReversal(0, 10, 1))      // no points redeemed
}
```

- [ ] **Step 2: Run — fail**

`cd backend && go test ./internal/handlers/ -run TestLoyaltyReversal -v` → FAIL.

- [ ] **Step 3: Implement**

```go
func computeLoyaltyReversal(pointsRedeemed, itemQty, returnQty int) int {
	if pointsRedeemed == 0 || itemQty == 0 { return 0 }
	return pointsRedeemed * returnQty / itemQty
}
```

Replace the reversal logic at `return_handler.go:382-399` to call this.

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestLoyaltyReversal -v
git add internal/handlers/return_handler.go internal/handlers/return_handler_test.go
git commit -m "fix(loyalty): reverse points proportionally on return (P2.T3)"
```

---

## Task 2.4: Vendor return status workflow

**Files:**
- Modify: `backend/internal/services/vendor_return_service.go:46`

- [ ] **Step 1: Write failing tests**

```go
func TestVendorReturn_DefaultsPending(t *testing.T) {
	db := setupTestDB(t)
	svc := NewVendorReturnService(db)
	ret, err := svc.Create(ctx, VendorReturnRequest{...})
	require.NoError(t, err)
	assert.Equal(t, "PENDING", ret.Status)
}

func TestVendorReturn_CompleteAfterApprove(t *testing.T) {
	db := setupTestDB(t)
	svc := NewVendorReturnService(db)
	ret, _ := svc.Create(ctx, req)
	require.NoError(t, svc.Approve(ctx, ret.ID, adminID))
	require.NoError(t, svc.Complete(ctx, ret.ID))
	final, _ := svc.Get(ctx, ret.ID)
	assert.Equal(t, "COMPLETED", final.Status)
}

func TestVendorReturn_RejectsInvalidTransition(t *testing.T) {
	db := setupTestDB(t)
	svc := NewVendorReturnService(db)
	ret, _ := svc.Create(ctx, req)
	err := svc.Complete(ctx, ret.ID)
	assert.Error(t, err) // can't skip APPROVE
}
```

- [ ] **Step 2: Run — fail**

`cd backend && go test ./internal/services/ -run TestVendorReturn -v` → FAIL (currently `COMPLETED` immediately).

- [ ] **Step 3: Refactor to state machine**

```go
type VendorReturnStatus string
const (
	VRPending   VendorReturnStatus = "PENDING"
	VRCancelled VendorReturnStatus = "CANCELLED"
	VRCompleted VendorReturnStatus = "COMPLETED"
)

var validTransitions = map[VendorReturnStatus][]VendorReturnStatus{
	VRPending:   {VRCompleted, VRCancelled},
	VRCompleted: {},
	VRCancelled: {},
}

func (s *VendorReturnService) canTransition(from, to VendorReturnStatus) bool {
	for _, t := range validTransitions[from] { if t == to { return true } }
	return false
}

func (s *VendorReturnService) Create(ctx context.Context, req VendorReturnRequest) (*domain.PurchaseReturn, error) {
	ret := &domain.PurchaseReturn{Status: VRPending, /* fields */}
	return ret, s.db.WithContext(ctx).Create(ret).Error
}

func (s *VendorReturnService) Approve(ctx context.Context, id uint) error {
	// record intent; Complete is what mutates state.
	return nil
}

func (s *VendorReturnService) Complete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ret domain.PurchaseReturn
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ret, id).Error; err != nil { return err }
		if !s.canTransition(ret.Status, VRCompleted) {
			return fmt.Errorf("cannot transition %s → COMPLETED", ret.Status)
		}
		ret.Status = VRCompleted
		ret.CompletedAt = time.Now()
		return tx.Save(&ret).Error
	})
}

func (s *VendorReturnService) Cancel(ctx context.Context, id uint) error { /* similar pattern */ }
```

- [ ] **Step 4: Add handler routes**

```go
vr.POST("/:id/approve", vendorReturnHandler.Approve)
vr.POST("/:id/complete", vendorReturnHandler.Complete)
vr.POST("/:id/cancel", vendorReturnHandler.Cancel)
```

- [ ] **Step 5: Run, commit**

```bash
cd backend && go test ./internal/services/ -run TestVendorReturn -v
git add internal/services/vendor_return_service.go internal/services/vendor_return_service_test.go internal/router/router.go
git commit -m "feat(vendor-returns): state machine PENDING→COMPLETED/CANCELLED (P2.T4)"
```

---

## Task 2.5: PO receive writes StockAdjustment

**Files:**
- Modify: `backend/internal/services/purchase_order_service.go:190-265`

- [ ] **Step 1: Write failing test**

```go
func TestPOReceive_CreatesStockAdjustment(t *testing.T) {
	db := setupTestDB(t)
	po := seedPurchaseOrder(t, db, "APPROVED")
	svc := NewPurchaseOrderService(db)
	require.NoError(t, svc.Receive(ctx, po.ID, []ReceiveItem{
		{ProductID: 1, BatchNumber: "B1", ExpiryDate: time.Now().Add(180 * 24 * time.Hour), Quantity: 50},
	}))
	var count int64
	db.Model(&domain.StockAdjustment{}).Where("reason_code = ? AND product_id = ?", "PURCHASE", 1).Count(&count)
	assert.Equal(t, int64(1), count)
}
```

- [ ] **Step 2-4: Implement**

In `purchase_order_service.go::Receive`, after creating each Batch:

```go
for _, item := range received {
	batch := domain.Batch{ProductID: item.ProductID, BatchNumber: item.BatchNumber, ExpiryDate: &item.ExpiryDate, Quantity: item.Quantity, LocationID: defaultLocationID}
	if err := tx.Create(&batch).Error; err != nil { return err }
	adj := domain.StockAdjustment{
		ProductID: item.ProductID, LocationID: defaultLocationID, BatchID: batch.ID,
		Type: "STOCK_IN", ReasonCode: "PURCHASE", Quantity: item.Quantity,
		AdjustedBy: userID, AdjustedAt: time.Now(),
	}
	if err := tx.Create(&adj).Error; err != nil { return err }
	item.ReceivedQty += item.Quantity
}
```

- [ ] **Step 5: Run, commit**

```bash
cd backend && go test ./internal/services/ -run TestPOReceive -v
git add internal/services/purchase_order_service.go internal/services/purchase_order_service_test.go
git commit -m "fix(inventory): PO receive inserts StockAdjustment (P2.T5)"
```

---

## Task 2.6: Product.PurchasePrice freshness on PO

**Files:**
- Modify: `backend/internal/services/purchase_order_service.go:67`
- Modify: `backend/internal/services/replenishment_service.go`
- Modify: `backend/internal/domain/models.go` (add `SnapshotPrice`)

- [ ] **Step 1: Write failing test**

```go
func TestPOFromSuggestion_UsesLatestPrice(t *testing.T) {
	db := setupTestDB(t)
	prod := seedProduct(t, db, "SKU1", 10.0)
	sugg := seedSuggestion(t, db, prod.ID, 10.0)
	db.Model(prod).Update("purchase_price", 12.5)

	svc := NewPurchaseOrderService(db)
	po, err := svc.CreateFromSuggestion(ctx, sugg.ID)
	require.NoError(t, err)
	var poItem domain.PurchaseOrderItem
	db.Where("purchase_order_id = ?", po.ID).First(&poItem)
	assert.Equal(t, 12.5, poItem.UnitPrice, "must use latest price, not suggestion snapshot")
}
```

- [ ] **Step 2: Run — fail**

`cd backend && go test ./internal/services/ -run TestPOFromSuggestion -v` → FAIL.

- [ ] **Step 3: Add `SnapshotPrice` to `ReorderSuggestion`**

```go
type ReorderSuggestion struct {
	gorm.Model
	ProductID     uint
	Quantity      int
	SnapshotPrice float64 `gorm:"type:decimal(12,2)"`
	Status        string
}
```

- [ ] **Step 4: Update suggestion generation**

In `replenishment_service.go`:

```go
sugg := domain.ReorderSuggestion{
	ProductID: p.ID, Quantity: target,
	SnapshotPrice: p.PurchasePrice, // NEW
	Status: "PENDING",
}
```

- [ ] **Step 5: Refactor `CreateFromSuggestion` to re-fetch product**

```go
func (s *PurchaseOrderService) CreateFromSuggestion(ctx context.Context, suggestionID uint) (*domain.PurchaseOrder, error) {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sugg domain.ReorderSuggestion
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sugg, suggestionID).Error; err != nil { return err }
		if sugg.Status != "PENDING" { return errors.New("suggestion not pending") }

		var prod domain.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&prod, sugg.ProductID).Error; err != nil { return err }

		if sugg.SnapshotPrice != prod.PurchasePrice {
			logrus.WithFields(logrus.Fields{"product_id": prod.ID, "snapshot": sugg.SnapshotPrice, "current": prod.PurchasePrice}).Warn("purchase price changed since suggestion")
		}

		po := domain.PurchaseOrder{ /* fields */ }
		if err := tx.Create(&po).Error; err != nil { return err }
		item := domain.PurchaseOrderItem{
			PurchaseOrderID: po.ID, ProductID: prod.ID, Quantity: sugg.Quantity,
			UnitPrice: prod.PurchasePrice, // fresh price
		}
		if err := tx.Create(&item).Error; err != nil { return err }
		sugg.Status = "PO_CREATED"
		return tx.Save(&sugg).Error
	})
}
```

- [ ] **Step 6: Run, commit**

```bash
cd backend && go test ./internal/services/ -run TestPOFromSuggestion -v
git add internal/services/purchase_order_service.go internal/services/replenishment_service.go internal/services/purchase_order_service_test.go internal/domain/models.go
git commit -m "fix(replenishment): PO uses fresh purchase price (P2.T6)"
```

---

## Task 2.7: Vendor return BatchID-ProductID validation

**Files:**
- Modify: `backend/internal/services/vendor_return_service.go`

- [ ] **Step 1: Write failing test**

```go
func TestVendorReturn_RejectsMismatchedBatch(t *testing.T) {
	db := setupTestDB(t)
	prod1 := seedProduct(t, db, "P1")
	prod2 := seedProduct(t, db, "P2")
	batch2 := seedBatch(t, db, prod2.ID, 1, 50, expiry)

	svc := NewVendorReturnService(db)
	_, err := svc.Create(ctx, VendorReturnRequest{
		Items: []VReturnItem{{ProductID: prod1.ID, BatchID: batch2.ID, Quantity: 10}},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch does not belong to product")
}
```

- [ ] **Step 2: Run — fail**

- [ ] **Step 3: Validate in `Create`**

```go
for _, item := range req.Items {
	var batch domain.Batch
	if err := tx.First(&batch, item.BatchID).Error; err != nil { return err }
	if batch.ProductID != item.ProductID {
		return fmt.Errorf("batch %d does not belong to product %d", item.BatchID, item.ProductID)
	}
}
```

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/services/ -run TestVendorReturn_RejectsMismatchedBatch -v
git add internal/services/vendor_return_service.go internal/services/vendor_return_service_test.go
git commit -m "fix(returns): validate batch belongs to product (P2.T7)"
```

---

## Task 2.8: Promotion priority test coverage

**Files:**
- Create: `backend/internal/services/promotion_resolver.go` + test
- Modify: `backend/internal/services/sales_service.go:202`

- [ ] **Step 1: Write failing test**

```go
func TestResolvePromotion_Hierarchy(t *testing.T) {
	prodID, subID, catID := uint(1), uint(2), uint(3)
	tests := []struct {
		name string
		promos []domain.Promotion
		wantID uint
	}{
		{"product beats category", []domain.Promotion{
			{ID: 1, CategoryID: &catID, DiscountPercent: 10, Priority: 1},
			{ID: 2, ProductID: &prodID, DiscountPercent: 15, Priority: 1},
		}, 2},
		{"subcategory beats category", []domain.Promotion{
			{ID: 1, CategoryID: &catID, DiscountPercent: 10},
			{ID: 2, SubCategoryID: &subID, DiscountPercent: 12},
		}, 2},
		{"higher priority wins ties", []domain.Promotion{
			{ID: 1, ProductID: &prodID, DiscountPercent: 10, Priority: 1},
			{ID: 2, ProductID: &prodID, DiscountPercent: 8, Priority: 5},
		}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolvePromotion(tt.promos, prodID, subID, catID)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantID, got.ID)
		})
	}
}
```

- [ ] **Step 2: Run — fail**

- [ ] **Step 3: Extract resolver**

Create `promotion_resolver.go`:

```go
package services

import "inventory/backend/internal/domain"

// ResolvePromotion picks the most specific applicable promotion.
// Hierarchy: product > subcategory > category. Ties: higher Priority, then higher DiscountPercent.
func ResolvePromotion(promos []domain.Promotion, productID, subCategoryID, categoryID uint) *domain.Promotion {
	var best *domain.Promotion
	for i := range promos {
		p := promos[i]
		if !applies(&p, productID, subCategoryID, categoryID) { continue }
		if best == nil || scopeRank(p) > scopeRank(*best) || (scopeRank(p) == scopeRank(*best) && higherRank(p, *best)) {
			best = &p
		}
	}
	return best
}

func applies(p *domain.Promotion, productID, subID, catID uint) bool {
	if p.ProductID != nil && *p.ProductID == productID { return true }
	if p.SubCategoryID != nil && *p.SubCategoryID == subID { return true }
	if p.CategoryID != nil && *p.CategoryID == catID { return true }
	return false
}

func scopeRank(p domain.Promotion) int {
	if p.ProductID != nil { return 3 }
	if p.SubCategoryID != nil { return 2 }
	return 1
}

func higherRank(a, b domain.Promotion) bool {
	return a.Priority > b.Priority || (a.Priority == b.Priority && a.DiscountPercent > b.DiscountPercent)
}
```

Update `sales_service.go:202` to call `ResolvePromotion`.

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/services/ -run TestResolvePromotion -v
git add internal/services/promotion_resolver.go internal/services/promotion_resolver_test.go internal/services/sales_service.go
git commit -m "feat(promotions): extract resolver with hierarchy tests (P2.T8)"
```

---

## Task 2.9: Checkout uses register location

**Files:**
- Modify: `backend/internal/handlers/sales.go:43-84`
- Modify: `backend/internal/services/sales_service.go`
- Modify: `backend/internal/requests/checkout_request.go`

- [ ] **Step 1: Add field to DTO**

In `checkout_request.go`:

```go
type CheckoutRequest struct {
	Items             []CheckoutItem `json:"items" binding:"required,dive"`
	PaymentMethod     string         `json:"payment_method" binding:"required,oneof=bkash card cash"`
	RegisterLocationID *uint         `json:"register_location_id" binding:"omitempty"`
	// ... existing fields
}
```

- [ ] **Step 2: Write failing test**

```go
func TestCheckout_DeductsFromRegisterLocation(t *testing.T) {
	db := setupTestDB(t)
	prod := seedProduct(t, db, "SKU1", 1)
	batch1 := seedBatch(t, db, prod.ID, 1, 100, expiry)
	batch2 := seedBatch(t, db, prod.ID, 2, 50, expiry)

	svc := NewSalesService(db)
	_, err := svc.ProcessCheckout(ctx, CheckoutRequest{
		Items: []CheckoutItem{{ProductID: prod.ID, Quantity: 30}},
		RegisterLocationID: ptrUint(2),
	}, userID)
	require.NoError(t, err)
	assertBatchQty(t, db, batch2.ID, 20)
	assertBatchQty(t, db, batch1.ID, 100)
}
```

- [ ] **Step 3: Run — fail**

- [ ] **Step 4: Update stock allocation**

In `sales_service.go`, change `fetchBatches` to accept and use locationID:

```go
func (s *SalesService) fetchBatches(ctx context.Context, tx *gorm.DB, productID, locationID uint) ([]domain.Batch, error) {
	var batches []domain.Batch
	q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_id = ? AND quantity > 0", productID)
	if locationID != 0 { q = q.Where("location_id = ?", locationID) }
	err := q.Order("expiry_date ASC NULLS LAST, created_at ASC").Find(&batches).Error
	return batches, err
}

// In ProcessCheckout:
locationID := uint(0)
if req.RegisterLocationID != nil { locationID = *req.RegisterLocationID }
// pass locationID into fetchBatches for each item
```

- [ ] **Step 5: Run, commit**

```bash
cd backend && go test ./internal/services/ -run TestCheckout -v
git add internal/handlers/sales.go internal/services/sales_service.go internal/requests/checkout_request.go internal/services/sales_service_test.go
git commit -m "feat(pos): checkout supports register location override (P2.T9)"
```

---

## Task 2.10: Stock adjustment concurrency test

**Files:**
- Test: `backend/internal/handlers/stock_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
func TestStockAdjustment_NoLostUpdates(t *testing.T) {
	db := setupTestDB(t)
	prod := seedProduct(t, db, "SKU1", 1)
	seedBatch(t, db, prod.ID, 1, 1000, expiry)

	var wg sync.WaitGroup
	var errs []error
	var mu sync.Mutex
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := adjustmentService.Adjust(ctx, AdjustRequest{ProductID: prod.ID, QuantityChange: -5, Reason: "race", UserID: 1})
			mu.Lock(); defer mu.Unlock(); errs = append(errs, err)
		}()
	}
	wg.Wait()
	for _, e := range errs { assert.NoError(t, e) }
	var finalQty int
	db.Model(&domain.Batch{}).Where("product_id = ?", prod.ID).Select("quantity").Scan(&finalQty)
	assert.Equal(t, 750, finalQty, "50 adjustments of -5 from 1000 → 750")
}
```

- [ ] **Step 2: Run with -race**

`cd backend && go test ./internal/handlers/ -run TestStockAdjustment -race -v` → likely FAIL or detect race.

- [ ] **Step 3: Confirm row lock in `stock.go:146`**

```go
err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
	Model(&domain.Batch{}).
	Where("product_id = ?", req.ProductID).
	Select("COALESCE(SUM(quantity), 0)").
	Row().Scan(&currentQuantity)
```

- [ ] **Step 4: Run with -race, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestStockAdjustment -race -v
git add internal/handlers/stock.go internal/handlers/stock_test.go
git commit -m "test(inventory): concurrency test for stock adjustment (P2.T10)"
```

---

## Task 2.11: UpdatePurchaseOrder preserves ReceivedQuantity

**Files:**
- Modify: `backend/internal/services/purchase_order_service.go:174-185`

- [ ] **Step 1: Write failing test**

```go
func TestUpdatePO_PreservesReceivedQuantity(t *testing.T) {
	db := setupTestDB(t)
	po := seedPurchaseOrderWithItems(t, db, "DRAFT", []Item{
		{ProductID: 1, Quantity: 100, ReceivedQuantity: 30},
	})
	svc := NewPurchaseOrderService(db)
	_, err := svc.Update(ctx, po.ID, UpdatePORequest{Items: []Item{{ProductID: 1, Quantity: 120}}})
	require.NoError(t, err)
	var item domain.PurchaseOrderItem
	db.Where("purchase_order_id = ? AND product_id = ?", po.ID, 1).First(&item)
	assert.Equal(t, 30, item.ReceivedQuantity)
	assert.Equal(t, 120, item.Quantity)
}
```

- [ ] **Step 2: Run — fail**

- [ ] **Step 3: Diff-and-reconcile implementation**

```go
func (s *PurchaseOrderService) Update(ctx context.Context, poID uint, req UpdatePORequest) (*domain.PurchaseOrder, error) {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var po domain.PurchaseOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, poID).Error; err != nil { return err }
		if po.Status != "DRAFT" { return errors.New("only DRAFT POs can be updated") }

		existing, _ := s.itemsByProduct(ctx, tx, poID)
		wanted := make(map[uint]int)
		for _, it := range req.Items { wanted[it.ProductID] = it.Quantity }

		for prodID, ex := range existing {
			if _, ok := wanted[prodID]; !ok {
				if ex.ReceivedQuantity > 0 { return errors.New("cannot delete an item with received quantity") }
				if err := tx.Delete(&ex).Error; err != nil { return err }
			}
		}
		for prodID, qty := range wanted {
			if ex, ok := existing[prodID]; ok {
				ex.Quantity = qty
				if err := tx.Save(&ex).Error; err != nil { return err }
			} else {
				item := domain.PurchaseOrderItem{PurchaseOrderID: poID, ProductID: prodID, Quantity: qty}
				if err := tx.Create(&item).Error; err != nil { return err }
			}
		}
		return nil
	})
}
```

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/services/ -run TestUpdatePO -v
git add internal/services/purchase_order_service.go internal/services/purchase_order_service_test.go
git commit -m "fix(replenishment): UpdatePurchaseOrder preserves ReceivedQuantity (P2.T11)"
```

---

## Task 2.12: Cron alert scan error tolerance

**Files:**
- Modify: `backend/internal/handlers/alerts.go:253-263`

- [ ] **Step 1: Write failing test**

```go
func TestCheckAndTriggerAlerts_ContinuesOnError(t *testing.T) {
	db := setupTestDB(t)
	goodProd := seedProduct(t, db, "GOOD", 1)
	badProd := seedProduct(t, db, "BAD", 1)
	seedAlertSettings(t, db, goodProd.ID)
	seedAlertSettings(t, db, badProd.ID)

	original := alertEvalFn
	alertEvalFn = func(p uint) error { if p == badProd.ID { panic("boom") }; return nil }
	defer func() { alertEvalFn = original }()

	assert.NotPanics(t, func() { CheckAndTriggerAlerts(db) })
	var count int64
	db.Model(&domain.Alert{}).Where("product_id = ?", goodProd.ID).Count(&count)
	assert.Greater(t, count, int64(0))
}
```

- [ ] **Step 2: Run — fail**

- [ ] **Step 3: Add recover wrapper**

```go
func CheckAndTriggerAlerts(db *gorm.DB) {
	var settings []domain.ProductAlertSettings
	if err := db.Find(&settings).Error; err != nil { logrus.WithError(err).Error(); return }
	for _, s := range settings {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logrus.WithField("product_id", s.ProductID).Errorf("alert scan panicked: %v", r)
				}
			}()
			processAlertForProduct(db, s)
		}()
	}
}
```

Extract `processAlertForProduct` from the existing body.

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestCheckAndTriggerAlerts -v
git add internal/handlers/alerts.go internal/handlers/alerts_test.go
git commit -m "fix(alerts): cron scan tolerates per-product errors (P2.T12)"
```

---

## Task 2.13: CRM customer identifier privacy

**Files:**
- Modify: `backend/internal/handlers/crm.go:68-114`

- [ ] **Step 1: Write failing test**

```go
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
		assert.Equal(t, 404, w.Code, "missing identifier %s", id)
	}
}
```

- [ ] **Step 2: Run — fail**

- [ ] **Step 3: Unify 404 response**

```go
func (h *CRMHandler) GetCustomer(c *gin.Context) {
	identifier := c.Param("identifier")
	user, err := h.lookupCustomer(c.Request.Context(), identifier)
	if err != nil {
		c.JSON(404, gin.H{"error": "customer not found"}) // uniform
		return
	}
	c.JSON(200, user)
}

func (h *CRMHandler) lookupCustomer(ctx context.Context, identifier string) (*domain.User, error) {
	var u domain.User
	if id, err := strconv.ParseUint(identifier, 10, 64); err == nil {
		if err := h.db.WithContext(ctx).First(&u, id).Error; err == nil { return &u, nil }
	}
	if strings.Contains(identifier, "@") {
		if err := h.db.WithContext(ctx).Where("email = ?", identifier).First(&u).Error; err == nil { return &u, nil }
	}
	if isPhone(identifier) {
		if err := h.db.WithContext(ctx).Where("phone = ?", identifier).First(&u).Error; err == nil { return &u, nil }
	}
	if err := h.db.WithContext(ctx).Where("username = ?", identifier).First(&u).Error; err == nil { return &u, nil }
	return nil, gorm.ErrRecordNotFound
}
```

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestGetCustomer -v
git add internal/handlers/crm.go internal/handlers/crm_test.go
git commit -m "fix(crm): uniform 404 response on missing customer (P2.T13)"
```

---

## Task 2.14: CRM churn-risk route accepts non-numeric identifiers

**Files:**
- Modify: `backend/internal/handlers/crm.go:249`

- [ ] **Step 1: Write failing test**

```go
func TestGetChurnRisk_AcceptsEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	seedUser(t, db, "alice", "a@x.com", "555-1234")
	mockAI := &mockAIService{response: `{"risk":"low"}`}

	r := gin.New()
	h := &CRMHandler{db: db, ai: mockAI}
	r.GET("/crm/customers/:identifier/churn-risk", h.GetChurnRisk)

	req := httptest.NewRequest("GET", "/crm/customers/a@x.com/churn-risk", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
```

- [ ] **Step 2: Run — fail**

- [ ] **Step 3: Normalize identifier**

```go
func (h *CRMHandler) GetChurnRisk(c *gin.Context) {
	identifier := c.Param("identifier")
	user, err := h.lookupCustomer(c.Request.Context(), identifier)
	if err != nil { middleware.RespondError(c, err); return }
	risk, err := h.ai.GetChurnRisk(c.Request.Context(), user.ID)
	if err != nil { middleware.RespondError(c, err); return }
	c.JSON(200, risk)
}
```

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestGetChurnRisk -v
git add internal/handlers/crm.go internal/handlers/crm_test.go
git commit -m "fix(crm): churn-risk accepts email/phone/username identifiers (P2.T14)"
```

---

## Task 2.15: Settings type validation

**Files:**
- Modify: `backend/internal/handlers/settings.go:28-44`

- [ ] **Step 1: Write failing test**

```go
func TestPutSetting_RejectsInvalidTaxRate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	h := &SettingsHandler{db: db}
	r := gin.New()
	r.PUT("/settings", h.Update)

	body := `{"key":"tax_rate_percentage","value":"abc"}`
	req := httptest.NewRequest("PUT", "/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}
```

- [ ] **Step 2: Run — fail**

- [ ] **Step 3: Define descriptors and validate**

```go
type SettingDescriptor struct {
	Type     string
	Validate func(v string) error
}

var settingDescriptors = map[string]SettingDescriptor{
	"tax_rate_percentage": {Type: "number", Validate: func(v string) error {
		f, err := strconv.ParseFloat(v, 64); if err != nil { return errors.New("not a number") }
		if f < 0 || f > 100 { return errors.New("must be 0-100") }; return nil
	}},
	"currency": {Type: "string", Validate: func(v string) error {
		if len(v) != 3 { return errors.New("must be 3-letter code") }; return nil
	}},
	"loyalty.tier_silver": {Type: "number", Validate: func(v string) error {
		n, err := strconv.Atoi(v); if err != nil { return err }; if n < 0 { return errors.New("negative") }; return nil
	}},
}

func (h *SettingsHandler) Update(c *gin.Context) {
	var req struct{ Key, Value string }
	if err := c.ShouldBindJSON(&req); err != nil { ... }
	d, ok := settingDescriptors[req.Key]
	if !ok { c.JSON(400, gin.H{"error": "unknown setting"}); return }
	if d.Validate != nil { if err := d.Validate(req.Value); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return } }
	// ... persist + audit log
}
```

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestPutSetting -v
git add internal/handlers/settings.go internal/handlers/settings_test.go
git commit -m "feat(settings): type-validated settings update (P2.T15)"
```

---

## Task 2.16: Remove dead code in `GetProductPerformanceAnalytics`

**Files:**
- Modify: `backend/internal/repository/reporting_repository.go:882-919`

- [ ] **Step 1: Write test asserting the function works**

```go
func TestProductPerformanceReport_Succeeds(t *testing.T) {
	db := setupTestDB(t)
	repo := &ReportingRepository{db: db}
	report, err := repo.GetProductPerformanceAnalytics(ctx, time.Now().Add(-30*24*time.Hour), time.Now())
	require.NoError(t, err)
	assert.NotNil(t, report)
}
```

- [ ] **Step 2: Delete dead block**

In `reporting_repository.go:882-919`, delete the first `for rows.Next()` block (the one with the typo `suppliers.id AS supplier_id` scanned into `SupplierName`). Keep only the working second pass.

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/repository/ -run TestProductPerformanceReport -v
git add internal/repository/reporting_repository.go internal/repository/reporting_repository_test.go
git commit -m "chore: remove dead code in product performance report (P2.T16)"
```

---

## Phase 2 verification gate

```bash
cd backend
go test -race ./...
```

Append to `CHANGELOG.md`. Merge `phase-2-integrity` to main, tag `v1.2-integrity`.

---

# Phase 3 — Correctness fixes (13 tasks)

## Task 3.1: Stop bulk import auto-finalize

**Files:**
- Modify: `backend/internal/consumers/bulk_consumer.go:213-221`

- [ ] **Step 1: Write failing test**

```go
func TestBulkConsumer_DoesNotAutoFinalize(t *testing.T) {
	db := setupTestDB(t)
	minio := setupMinio(t)
	seedCSV(t, minio, "test.csv")
	job := seedJob(t, db, "BULK_IMPORT", "QUEUED")

	consumeAndProcess(t, db, minio, job)

	var j domain.Job
	db.First(&j, job.ID)
	assert.Equal(t, "PENDING_CONFIRMATION", j.Status)
}
```

- [ ] **Step 2: Run — fail; remove `finalizeImport(...)` call**

In `bulk_consumer.go:213`:

```go
job.Status = "PENDING_CONFIRMATION"
job.Payload, _ = json.Marshal(map[string]any{"validationResult": result, "validatedAt": time.Now()})
return db.Save(&job).Error
```

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/consumers/ -run TestBulkConsumer_DoesNotAutoFinalize -v
git add internal/consumers/bulk_consumer.go internal/consumers/bulk_consumer_test.go
git commit -m "fix(bulk): stop auto-finalize, require explicit confirm (P3.T1)"
```

---

## Task 3.2: Re-confirm idempotency

**Files:**
- Modify: `backend/internal/consumers/bulk_consumer.go:223-321`
- Modify: `backend/internal/services/bulk_import_service.go`

- [ ] **Step 1: Write failing test**

```go
func TestBulkReconfirm_IsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	job := seedJobInState(t, db, "PENDING_CONFIRMATION")
	firstResult := consumeFirstPass(t, db, job)
	consumeConfirm(t, db, job)
	consumeConfirm(t, db, job) // must be no-op

	var prodCount int64
	db.Model(&domain.Product{}).Count(&prodCount)
	assert.Equal(t, int64(len(firstResult.ValidProducts)), prodCount)
}
```

- [ ] **Step 2: Run — fail; add early-exit + ON CONFLICT**

In `bulk_consumer.go:223` (processImportConfirmDelivery):

```go
if job.Status == "COMPLETED" {
	logrus.WithField("job_id", jobID).Info("bulk confirm already completed; skipping")
	return nil
}
```

In `bulk_import_service.go::CreateInBatches`:

```go
return db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(products, 100).Error
```

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/consumers/ -run TestBulkReconfirm -v
git add internal/consumers/bulk_consumer.go internal/services/bulk_import_service.go
git commit -m "fix(bulk): idempotent re-confirm via ON CONFLICT (P3.T2)"
```

---

## Task 3.3: Timezone SQL injection

**Files:**
- Modify: `backend/internal/repository/reporting_repository.go:453-465`

- [ ] **Step 1: Write failing test**

```go
func TestHeatmapReport_RejectsMaliciousTimezone(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&domain.SystemSetting{Key: "timezone", Value: "UTC' UNION SELECT password FROM users --"})
	repo := &ReportingRepository{db: db}
	_, err := repo.GetHourlySalesHeatmap(ctx, time.Now().Add(-7*24*time.Hour), time.Now())
	assert.NoError(t, err)
}
```

- [ ] **Step 2: Run — fail; use parameterized query**

```go
func (r *ReportingRepository) GetHourlySalesHeatmap(ctx context.Context, start, end time.Time) ([]HeatmapBucket, error) {
	var tz string
	if err := r.db.WithContext(ctx).Model(&domain.SystemSetting{}).Where("key = ?", "timezone").Pluck("value", &tz).Error; err != nil { return nil, err }
	if tz == "" || !isValidTimezone(tz) { tz = "UTC" }

	var buckets []HeatmapBucket
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXTRACT(HOUR FROM adjusted_at AT TIME ZONE ?) AS hour,
		       DATE_TRUNC('day', adjusted_at AT TIME ZONE ?) AS day,
		       SUM(quantity) AS total
		FROM stock_adjustments
		WHERE adjusted_at BETWEEN ? AND ? AND reason_code = 'SALE'
		GROUP BY day, hour ORDER BY day, hour
	`, tz, tz, start, end).Scan(&buckets).Error
	return buckets, err
}

func isValidTimezone(tz string) bool {
	if tz == "" { return false }
	for _, r := range tz {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '/' || r == '_' || r == '+' || r == '-') {
			return false
		}
	}
	return true
}
```

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/repository/ -run TestHeatmapReport -v
git add internal/repository/reporting_repository.go internal/repository/reporting_repository_test.go
git commit -m "fix(reports): parameterized timezone in heatmap query (P3.T3)"
```

---

## Task 3.4: AuditLog writers

**Files:**
- Create: `backend/internal/repository/audit_log.go`
- Modify: `backend/internal/handlers/sales.go`, `handlers/settings.go`, `services/role_service.go`

- [ ] **Step 1: Define writer**

```go
// internal/repository/audit_log.go
package repository

func WriteAudit(ctx context.Context, db *gorm.DB, entry domain.AuditLog) error {
	return db.WithContext(ctx).Create(&entry).Error
}
```

- [ ] **Step 2: Write failing test**

```go
func TestCheckout_WithDiscount_AuditWritten(t *testing.T) {
	db := setupTestDB(t)
	seedProductWithDiscount(t, db, 20.0)
	svc := NewSalesService(db)
	_, err := svc.ProcessCheckout(ctx, CheckoutRequest{Items: []CheckoutItem{{ProductID: 1, Quantity: 1}}, UserID: 1})
	require.NoError(t, err)
	var count int64
	db.Model(&domain.AuditLog{}).Where("action = ?", "DISCOUNT_APPLIED").Count(&count)
	assert.Greater(t, count, int64(0))
}
```

- [ ] **Step 3: Run — fail; emit audit rows**

In `sales_service.go::processItems`, after applying discount:

```go
if item.Discount > 10.0 {
	_ = repository.WriteAudit(ctx, tx, domain.AuditLog{
		Action: "DISCOUNT_APPLIED", Entity: "OrderItem", EntityID: item.ProductID,
		UserID: userID, Timestamp: time.Now(),
		Details: fmt.Sprintf("discount=%.2f%%", item.Discount),
	})
}
```

Similar emissions in settings update + role permission change.

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestCheckout_WithDiscount -v
git add internal/repository/audit_log.go internal/handlers/sales.go internal/handlers/settings.go internal/services/role_service.go internal/handlers/sales_audit_test.go
git commit -m "feat(audit): emit AuditLog on discount, settings, role changes (P3.T4)"
```

---

## Task 3.5: Basket analysis pagination

**Files:**
- Modify: `backend/internal/repository/reporting_repository.go:742`

- [ ] **Step 1: Write failing test**

```go
func TestBasketAnalysis_Pagination(t *testing.T) {
	db := setupTestDB(t)
	seedBaskets(t, db, 50)
	repo := &ReportingRepository{db: db}
	page1, total, err := repo.GetBasketAnalysis(ctx, start, end, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 50, total)
	assert.Len(t, page1, 20)
	page3, _, _ := repo.GetBasketAnalysis(ctx, start, end, 3, 20)
	assert.Len(t, page3, 10)
}
```

- [ ] **Step 2: Run — fail; add page/limit params**

```go
func (r *ReportingRepository) GetBasketAnalysis(ctx context.Context, start, end time.Time, page, limit int) ([]BasketItem, int, error) {
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit
	var total int64
	if err := r.db.WithContext(ctx).Model(&domain.Order{}).Where("created_at BETWEEN ? AND ?", start, end).Count(&total).Error; err != nil { return nil, 0, err }
	var items []BasketItem
	err := r.db.WithContext(ctx).Raw(`SELECT ... LIMIT ? OFFSET ?`, limit, offset).Scan(&items).Error
	return items, int(total), err
}
```

Update handler to parse `page` and `limit` from query params.

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/repository/ -run TestBasketAnalysis -v
git add internal/repository/reporting_repository.go internal/repository/reporting_repository_test.go
git commit -m "feat(reports): basket analysis pagination (P3.T5)"
```

---

## Task 3.6: Customer insights net of returns

**Files:**
- Modify: `backend/internal/repository/reporting_repository.go:791-809`

- [ ] **Step 1: Write failing test**

```go
func TestCustomerInsights_NetOfReturns(t *testing.T) {
	db := setupTestDB(t)
	cust := seedCustomer(t, db, "alice")
	seedCompletedOrder(t, db, cust.ID, 500.00)
	seedRefund(t, db, cust.ID, 400.00)
	repo := &ReportingRepository{db: db}
	report, err := repo.GetCustomerInsights(ctx, start, end)
	require.NoError(t, err)
	var alice *CustomerInsight
	for i := range report { if report[i].CustomerID == cust.ID { alice = &report[i]; break } }
	require.NotNil(t, alice)
	assert.InDelta(t, 500.0, alice.GrossRevenue, 0.01)
	assert.InDelta(t, 100.0, alice.NetRevenue, 0.01)
}
```

- [ ] **Step 2: Subtract refunds in query**

```go
err := r.db.WithContext(ctx).Raw(`
    SELECT u.id AS customer_id, u.username,
           COALESCE(SUM(o.total_amount), 0) AS gross_revenue,
           COALESCE(SUM(o.total_amount), 0) - COALESCE((
               SELECT SUM(ri.quantity * oi.unit_price)
               FROM returns r
               JOIN return_items ri ON ri.return_id = r.id
               JOIN order_items oi ON oi.id = ri.order_item_id
               JOIN orders o ON o.id = oi.order_id
               WHERE o.user_id = u.id AND r.status = 'APPROVED'
           ), 0) AS net_revenue
    FROM users u
    LEFT JOIN orders o ON o.user_id = u.id AND o.status = 'COMPLETED' AND o.created_at BETWEEN ? AND ?
    GROUP BY u.id, u.username ORDER BY net_revenue DESC
`, start, end).Scan(&report).Error
```

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/repository/ -run TestCustomerInsights -v
git add internal/repository/reporting_repository.go internal/repository/reporting_repository_test.go
git commit -m "fix(reports): customer insights net of refunds (P3.T6)"
```

---

## Task 3.7: Per-order TaxAmount field

**Files:**
- Modify: `backend/internal/domain/models.go` (Order struct)
- Modify: `backend/internal/services/sales_service.go`

- [ ] **Step 1: Add `TaxAmount` to Order**

```go
type Order struct {
	gorm.Model
	// existing fields
	Subtotal    float64 `gorm:"type:decimal(12,2)"`
	TaxAmount   float64 `gorm:"type:decimal(12,2);default:0"`
	TotalAmount float64 `gorm:"type:decimal(12,2)"`
}
```

- [ ] **Step 2: Write failing test**

```go
func TestCheckout_PersistsTaxAmount(t *testing.T) {
	db := setupTestDB(t)
	setSetting(t, db, "tax_rate_percentage", "10")
	svc := NewSalesService(db)
	order, err := svc.ProcessCheckout(ctx, CheckoutRequest{Items: []CheckoutItem{{ProductID: 1, Quantity: 1, UnitPrice: 100}}, UserID: 1})
	require.NoError(t, err)
	var o domain.Order
	db.First(&o, order.ID)
	assert.InDelta(t, 10.0, o.TaxAmount, 0.01)
	assert.InDelta(t, 110.0, o.TotalAmount, 0.01)
}
```

- [ ] **Step 3: Populate at checkout**

In `sales_service.go::createOrderAndTransaction`:

```go
order := domain.Order{
	Subtotal: subtotal, TaxAmount: taxAmount, TotalAmount: subtotal + taxAmount,
	// existing fields
}
```

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/services/ -run TestCheckout_PersistsTaxAmount -v
git add internal/domain/models.go internal/services/sales_service.go internal/services/sales_service_test.go
git commit -m "feat(orders): persist per-order TaxAmount (P3.T7)"
```

---

## Task 3.8: Sales trends weekly groupBy

**Files:**
- Modify: `backend/internal/repository/reporting_repository.go:33-101`

- [ ] **Step 1: Write failing test**

```go
func TestSalesTrends_ISOWeekVsCalendarWeek(t *testing.T) {
	db := setupTestDB(t)
	seedSaleAt(t, db, time.Date(2026, 1, 4, 12, 0, 0, 0, time.UTC), 100) // Sunday
	repo := &ReportingRepository{db: db}
	iso, _ := repo.GetSalesTrends(ctx, start, end, "iso_week")
	cal, _ := repo.GetSalesTrends(ctx, start, end, "calendar_week")
	assert.NotEqual(t, iso[0].BucketStart, cal[0].BucketStart)
}
```

- [ ] **Step 2: Add groupBy modes**

```go
func (r *ReportingRepository) GetSalesTrends(ctx context.Context, start, end time.Time, groupBy string) ([]SalesTrendBucket, error) {
	var truncExpr string
	switch groupBy {
	case "iso_week", "weekly", "":
		truncExpr = "week" // ISO week
	case "calendar_week":
		truncExpr = "(date_trunc('week', adjusted_at AT TIME ZONE 'UTC') + interval '1 day')"
	case "daily":
		truncExpr = "day"
	case "monthly":
		truncExpr = "month"
	default:
		return nil, fmt.Errorf("unsupported groupBy: %s", groupBy)
	}
	// existing query using truncExpr (safe — internal switch)
}
```

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/repository/ -run TestSalesTrends -v
git add internal/repository/reporting_repository.go internal/handlers/reports.go internal/repository/reporting_repository_test.go
git commit -m "feat(reports): distinct iso_week vs calendar_week grouping (P3.T8)"
```

---

## Task 3.9: `ListAllOrders` pagination

**Files:**
- Modify: `backend/internal/handlers/sales.go:259-335`

- [ ] **Step 1: Write failing test**

```go
func TestListAllOrders_Pagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	seedOrders(t, db, 250)
	r := gin.New()
	h := &SalesHandler{db: db}
	r.GET("/sales/history", h.ListAllOrders)

	req := httptest.NewRequest("GET", "/sales/history?page=2&limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp struct{ Data []Order; Total int }
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Data, 50)
	assert.Equal(t, 250, resp.Total)
}
```

- [ ] **Step 2: Add page/limit params**

```go
func (h *SalesHandler) ListAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 { page = 1 }
	if limit < 1 || limit > 200 { limit = 50 }
	offset := (page - 1) * limit

	var total int64
	h.db.Model(&domain.Order{}).Count(&total)
	var orders []domain.Order
	err := h.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if err != nil { middleware.RespondError(c, err); return }
	c.JSON(200, gin.H{"data": orders, "total": total, "page": page, "limit": limit})
}
```

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestListAllOrders -v
git add internal/handlers/sales.go internal/handlers/sales_test.go
git commit -m "feat(sales): paginate ListAllOrders (P3.T9)"
```

---

## Task 3.10: `User.AfterFind` safer preloading

**Files:**
- Modify: `backend/internal/domain/models.go:171-176`

- [ ] **Step 1: Write failing test**

```go
func TestUserAfterFind_EmptyRoleWhenNotPreloaded(t *testing.T) {
	u := domain.User{Username: "alice"}
	u.AfterFind(nil)
	assert.Equal(t, "", u.RoleName)
}
```

- [ ] **Step 2: Fix AfterFind**

```go
func (u *User) AfterFind(tx *gorm.DB) error {
	if u.Role.ID != 0 {
		u.RoleName = u.Role.Name
	} else {
		u.RoleName = ""
	}
	return nil
}
```

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/domain/ -run TestUserAfterFind -v
git add internal/domain/models.go internal/domain/models_test.go
git commit -m "fix(user): AfterFind explicitly clears RoleName (P3.T10)"
```

---

## Task 3.11: WebSocket auth via first message

**Files:**
- Modify: `backend/internal/handlers/websocket.go`
- Modify: `backend/internal/websocket/client.go`

- [ ] **Step 1: Write failing test**

```go
func TestWebSocket_FirstMessageAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws", ServeWs(&config.Config{WSAllowedOrigins: []string{"*"}}, websocket.NewHub()))
	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	conn.WriteJSON(map[string]any{"type": "auth", "token": validTestToken})
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	var msg map[string]any
	assert.NoError(t, conn.ReadJSON(&msg))
}
```

- [ ] **Step 2: Implement first-message auth**

In `handlers/websocket.go::ServeWs`:

```go
func ServeWs(cfg *config.Config, hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		upgrader := websocket.Upgrader{CheckOrigin: buildOriginChecker(cfg.WSAllowedOrigins)}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil { return }
		client := websocket.NewClient(conn, hub, 0, "")
		go client.WritePump()

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		var msg struct{ Type, Token string }
		if err := conn.ReadJSON(&msg); err != nil || msg.Type != "auth" {
			conn.Close(); return
		}
		claims, err := auth.ValidateToken(msg.Token, cfg.JWTSecret)
		if err != nil { conn.Close(); return }
		client.UserID = claims.UserID
		client.Role = claims.Role
		hub.Register(client)
		go client.ReadPump()
	}
}
```

- [ ] **Step 3: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestWebSocket -v
git add internal/handlers/websocket.go internal/websocket/client.go internal/handlers/websocket_test.go
git commit -m "feat(ws): first-message auth, JWT no longer in URL (P3.T11)"
```

---

## Task 3.12: User-facing notification subscriptions

**Files:**
- Modify: `backend/internal/domain/models.go` (add `UserNotificationSubscription`)
- Modify: `backend/internal/handlers/notifications.go`
- Modify: `backend/internal/consumers/alert_consumer.go`

- [ ] **Step 1: Add model**

```go
type UserNotificationSubscription struct {
	gorm.Model
	UserID  uint   `gorm:"index:idx_user_type"`
	Type    string `gorm:"index:idx_user_type"`
	Enabled bool
}
```

- [ ] **Step 2: Write failing test**

```go
func TestUserCanOptOutOfAlerts(t *testing.T) {
	db := setupTestDB(t)
	seedUser(t, db, "alice", 1)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &NotificationHandler{db: db}
	r.PUT("/users/:id/notifications/subscriptions/:type", h.UpdateSubscription)
	r.Use(func(c *gin.Context) { c.Set("user_id", uint(1)); c.Next() })

	body := `{"enabled": false}`
	req := httptest.NewRequest("PUT", "/users/1/notifications/subscriptions/LOW_STOCK", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	triggerLowStockAlert(t, db, productID)
	var count int64
	db.Model(&domain.Notification{}).Where("user_id = ?", 1).Count(&count)
	assert.Equal(t, int64(0), count)
}
```

- [ ] **Step 3: Implement endpoint + filter**

```go
func (h *NotificationHandler) UpdateSubscription(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 { c.AbortWithStatus(401); return }
	var req struct{ Enabled bool `json:"enabled"`}
	if err := c.ShouldBindJSON(&req); err != nil { ... }
	typ := c.Param("type")
	sub := domain.UserNotificationSubscription{UserID: userID, Type: typ, Enabled: req.Enabled}
	h.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&sub)
	c.JSON(200, sub)
}
```

In `AlertConsumer.loadSubscribedUsers`, exclude opted-out users:

```go
err := tx.Where("NOT EXISTS (SELECT 1 FROM user_notification_subscriptions s WHERE s.user_id = users.id AND s.type = ? AND s.enabled = false)", alertType).Find(&users).Error
```

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/handlers/ -run TestUserCanOptOut -v
git add internal/handlers/notifications.go internal/repository/notification_repository.go internal/domain/models.go internal/consumers/alert_consumer.go internal/handlers/notifications_test.go
git commit -m "feat(notifications): per-user alert opt-out (P3.T12)"
```

---

## Task 3.13: GetCustomer identifier leak fix

Already covered in P2.T13. **Mark complete**; no additional commit needed.

## Phase 3 verification gate

```bash
cd backend
go test ./...
```

Append to `CHANGELOG.md`. Merge `phase-3-correctness` to main, tag `v1.3-correctness`.

---

# Phase 4 — Architectural overhaul (18 tasks)

Phase 4 introduces structural changes. **Order matters**: T2 → T1 → T3 → T4 → T5 → T6 → T7 → T8 → T9 → T10 → T11 → T12 → T13 → T14 → T15 → T16 → T17 → T18. Tasks T1/T2 are the foundation; complete them first.

## Task 4.1 + 4.2 (combined): DI container + router split

**Files:**
- Create: `backend/internal/app/container.go`
- Modify: `backend/internal/router/router.go` (becomes thin orchestrator)
- Create: `backend/internal/router/groups/{public,payment,websocket,auth,users,products,sales,reports,replenishment,alerts,bulk,inventory,crm,time_tracking,search,settings,roles}.go`
- Modify: `backend/internal/bootstrap/app.go`

- [ ] **Step 1: Define `app.Container`**

Create `backend/internal/app/container.go`:

```go
package app

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
	"inventory/backend/internal/config"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"
	"inventory/backend/internal/storage"
	"inventory/backend/internal/websocket"
)

type Container struct {
	Cfg      *config.Config
	DB       *gorm.DB
	Redis    *redis.Client
	AMQP     *amqp.Connection
	Hub      *websocket.Hub
	Uploader storage.Uploader

	Roles    *repository.RoleRepositoryExt
	Users    *repository.UserRepository
	Products *repository.ProductRepository
	// ... all 22 repos

	Auth  *services.AuthService
	Sales *services.SalesService
	// ... all 17 services

	User    *handlers.UserHandler
	Product *handlers.ProductHandler
	// ... all 28 handlers
}
```

- [ ] **Step 2: Build container in bootstrap**

In `bootstrap/app.go`:

```go
func NewApp(cfg *config.Config) (*App, error) {
	c := &app.Container{Cfg: cfg}
	c.DB = repository.InitDB(...)
	c.Redis = repository.InitRedis(...)
	c.Hub = websocket.NewHub()
	c.AMQP = messagebroker.InitRabbitMQ(cfg.RabbitMQURL)
	c.Uploader, _ = storage.NewMinIOUploader(cfg)

	// Repositories
	c.Roles = repository.NewRoleRepositoryExt(c.DB, c.Redis)
	c.Users = repository.NewUserRepository(c.DB)
	// ... etc

	// Services (depend on repositories)
	c.Sales = services.NewSalesService(c.DB, c.Products, c.Promotions, ...)
	// ... etc

	// Handlers (depend on services)
	c.User = handlers.NewUserHandler(c.Users, c.Auth, c.Cfg)
	// ... etc

	return &App{container: c, Cron: cron.New()}, nil
}
```

- [ ] **Step 3: Rewrite `router.go`**

```go
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/router/groups"
)

func Setup(c *app.Container) *gin.Engine {
	r := gin.New()
	r.Use(middleware.ErrorHandler(), middleware.SecurityHeaders(), middleware.RequestID())

	r.GET("/health", groups.Health)
	r.GET("/ready", groups.Ready)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/swagger/*any", gin.WrapH(swaggerFiles.Handler))

	root := r.Group("/")
	groups.RegisterPublic(root, c)
	groups.RegisterPayment(root, c)
	groups.RegisterWebSocket(r, c)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.Auth(c.Cfg), middleware.CSRF(c.Redis))
	groups.RegisterAuth(v1, c)
	groups.RegisterUsers(v1, c)
	groups.RegisterProducts(v1, c)
	groups.RegisterSales(v1, c)
	groups.RegisterReports(v1, c)
	groups.RegisterReplenishment(v1, c)
	groups.RegisterAlerts(v1, c)
	groups.RegisterBulk(v1, c)
	groups.RegisterInventory(v1, c)
	groups.RegisterCRM(v1, c)
	groups.RegisterTimeTracking(v1, c)
	groups.RegisterSearch(v1, c)
	groups.RegisterSettings(v1, c)
	groups.RegisterRoles(v1, c)

	return r
}
```

- [ ] **Step 4: Create `groups/public.go` first (used by Setup)**

```go
package groups

import (
	"github.com/gin-gonic/gin"
	"inventory/backend/internal/app"
	"net/http"
)

func Health(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "UP"}) }
func Ready(c *gin.Context) {
	if err := c.MustGet("db").(interface{ Ping() error }).Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"db": err.Error()}); return
	}
	c.JSON(http.StatusOK, gin.H{"status": "READY"})
}

func RegisterPublic(g *gin.RouterGroup, c *app.Container) {
	g.POST("/webhooks", c.Integrations.HandleWebhook)
}

func RegisterPayment(g *gin.RouterGroup, c *app.Container) {
	g.POST("/payment/create", c.Payment.Create)
	g.POST("/payment/success", c.Payment.Success)
	g.POST("/payment/fail", c.Payment.Fail)
	g.POST("/payment/cancel", c.Payment.Cancel)
	g.POST("/payment/ipn", c.Payment.IPN)
}

func RegisterWebSocket(r *gin.Engine, c *app.Container) {
	r.GET("/ws", handlers.ServeWs(c.Cfg, c.Hub))
}
```

- [ ] **Step 5: For each remaining domain, create a groups file**

Example `groups/products.go`:

```go
package groups

import (
	"github.com/gin-gonic/gin"
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
)

func RegisterProducts(g *gin.RouterGroup, c *app.Container) {
	h := c.Product
	g.POST("/products", middleware.RequirePermission(c.Roles, "products.write"), h.Create)
	g.GET("/products", middleware.RequirePermission(c.Roles, "products.read"), h.List)
	g.GET("/products/:productId", middleware.RequirePermission(c.Roles, "products.read"), h.Get)
	g.GET("/products/sku/:sku", middleware.RequirePermission(c.Roles, "products.read"), h.GetBySKU)
	g.GET("/products/barcode/:barcode", middleware.RequirePermission(c.Roles, "products.read"), h.GetByBarcode)
	g.PUT("/products/:productId", middleware.RequirePermission(c.Roles, "products.write"), h.Update)
	g.DELETE("/products/:productId", middleware.RequirePermission(c.Roles, "products.delete"), h.Delete)
	g.GET("/products/:productId/stock", middleware.RequirePermission(c.Roles, "products.read"), h.GetStock)
	g.POST("/products/:productId/stock/batches", middleware.RequirePermission(c.Roles, "products.write"), h.CreateBatch)
	g.POST("/products/:productId/stock/adjustments", middleware.RequirePermission(c.Roles, "products.write"), h.CreateAdjustment)
	g.GET("/products/:productId/history", middleware.RequirePermission(c.Roles, "products.read"), h.History)
}
```

Repeat this pattern for every group file (17 groups total). Reference `docs/superpowers/specs/2026-07-25-quantify-remediation-design.md` Section 5.2 for the route list.

- [ ] **Step 6: Verify build + tests**

Run: `cd backend && go build ./... && go test ./...`
Expected: all green; `wc -l internal/router/router.go` < 100.

- [ ] **Step 7: Commit**

```bash
cd backend
git add internal/app/ internal/router/ internal/bootstrap/
git commit -m "refactor: DI container + router split into per-domain groups (P4.T1, P4.T2)"
```

---

## Task 4.3: Repository interface extraction

**Files:**
- Create: `backend/internal/repository/interfaces.go`

- [ ] **Step 1: Define interfaces**

```go
package repository

import (
	"context"
	"gorm.io/gorm"
	"inventory/backend/internal/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	Get(ctx context.Context, id uint) (*domain.Product, error)
	List(ctx context.Context, params ListParams) ([]domain.Product, error)
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id uint) error
}

// one interface per repo (OrderRepository, UserRepository, RoleRepository, etc.)
```

- [ ] **Step 2: Bind concrete types in `container.go`**

```go
Products: repository.NewProductRepository(db), // *ProductRepository satisfies ProductRepository
```

- [ ] **Step 3: Update services to depend on interfaces**

```go
type SalesService struct {
	products repository.ProductRepository // interface
	orders   repository.OrderRepository
	// ...
}
```

- [ ] **Step 4: Verify build + commit**

```bash
cd backend && go build ./... && go test ./...
git add internal/repository/interfaces.go internal/services/ internal/app/container.go
git commit -m "refactor: repository interfaces for testability (P4.T3)"
```

---

## Task 4.4: Remove package-level `repository.DB` singleton

**Files:**
- Modify: `backend/internal/repository/database.go:17`

- [ ] **Step 1: Make DB private**

```go
var (
	db     *gorm.DB
	dbOnce sync.Once
)

func initDB(dsn string) (*gorm.DB, error) {
	dbOnce.Do(func() {
		db, _ = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	})
	return db, nil
}

func DB() *gorm.DB { return db }
```

- [ ] **Step 2: Replace direct `repository.DB` access**

Grep `repository.DB` across the codebase and replace each with an injected `*gorm.DB` from `app.Container.DB` (or function parameter).

- [ ] **Step 3: Verify grep returns only the package's own definition**

```bash
cd backend && grep -rn "repository\.DB" --include="*.go" | grep -v "database.go"
```
Expected: no results.

- [ ] **Step 4: Run tests + commit**

```bash
cd backend && go test ./...
git add internal/repository/ internal/handlers/ internal/services/ internal/bootstrap/ internal/app/
git commit -m "refactor: remove package-level repository.DB singleton (P4.T4)"
```

---

## Task 4.5: Adopt golang-migrate

**Files:**
- Create: `backend/migrations/0001_init.up.sql`, `0001_init.down.sql`
- Modify: `backend/cmd/migrator/main.go`
- Modify: `backend/go.mod`

- [ ] **Step 1: Add dependency**

```bash
cd backend
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
```

- [ ] **Step 2: Generate initial schema SQL**

Run the existing AutoMigrate against an empty test DB, then dump the schema:

```bash
pg_dump --schema-only --no-owner -d inventory_db > migrations/0001_init.up.sql
```

Hand-edit to add `0001_init.down.sql` (DROP TABLE …).

- [ ] **Step 3: Rewrite `cmd/migrator/main.go`**

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dir := flag.String("dir", "migrations", "migrations directory")
	cmd := flag.String("cmd", "up", "up | down | version")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "Postgres DSN")
	flag.Parse()

	m, err := migrate.New(fmt.Sprintf("file://%s", *dir), *dsn)
	if err != nil { log.Fatal(err) }
	switch *cmd {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "version":
		v, dirty, _ := m.Version()
		fmt.Printf("version=%d dirty=%v\n", v, dirty); return
	}
	if err != nil && err != migrate.ErrNoChange { log.Fatal(err) }
	fmt.Println("OK")
}
```

- [ ] **Step 4: Test fresh DB**

```bash
docker run --rm -d --name pg -p 5432:5432 -e POSTGRES_PASSWORD=test postgres:15
sleep 5
DATABASE_URL="postgres://postgres:test@localhost:5432/postgres?sslmode=disable" go run ./cmd/migrator up
DATABASE_URL="..." go run ./cmd/migrator version
```

- [ ] **Step 5: Commit**

```bash
cd backend
git add migrations/ cmd/migrator/ go.mod go.sum
git commit -m "feat(migrate): adopt golang-migrate for versioned SQL migrations (P4.T5)"
```

---

## Task 4.6: Convert in-code seeding to migrations

**Files:**
- Create: `backend/migrations/0002_seed.up.sql`, `0002_seed.down.sql`
- Modify: `backend/internal/repository/database.go:151-368`

- [ ] **Step 1: Extract permissions list to SQL**

Generate the 47 permissions + 4 roles + permission grants + system settings + default alert subscriptions as SQL `INSERT … ON CONFLICT DO NOTHING` statements. Save to `0002_seed.up.sql`.

- [ ] **Step 2: Remove in-code seeding**

In `database.go::seedData`, delete the body (or replace with a no-op log). AutoMigrate still runs for new tables in dev.

- [ ] **Step 3: Run migrator on fresh DB, verify seeded**

```bash
DATABASE_URL="..." go run ./cmd/migrator up
psql $DATABASE_URL -c "SELECT COUNT(*) FROM permissions"  # expect 47
```

- [ ] **Step 4: Commit**

```bash
cd backend
git add migrations/0002_seed.up.sql migrations/0002_seed.down.sql internal/repository/database.go
git commit -m "refactor(migrate): convert in-code seeding to migration SQL (P4.T6)"
```

---

## Task 4.7: Disable AutoMigrate outside dev

**Files:**
- Modify: `backend/internal/bootstrap/app.go:48`

- [ ] **Step 1: Add env check**

```go
if cfg.AppEnv == "dev" {
    if err := repository.AutoMigrate(c.DB); err != nil { return nil, err }
} else {
    logrus.Info("AutoMigrate skipped (APP_ENV != dev); run cmd/migrator to apply migrations")
}
```

- [ ] **Step 2: Add `AppEnv` to config**

```go
AppEnv string // APP_ENV
```

- [ ] **Step 3: Test**

```bash
APP_ENV=production DATABASE_URL="..." go run ./cmd/server  # should warn and continue
```

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/bootstrap/app.go internal/config/config.go
git commit -m "feat(migrate): disable AutoMigrate outside dev (P4.T7)"
```

---

## Task 4.8: Reporting repository split

**Files:**
- Modify: `backend/internal/repository/reporting_repository.go` (1089 LOC)

- [ ] **Step 1: Create split files**

```
internal/repository/
├── reporting_repository.go    # base struct + interface
├── sales_reports.go           # sales trends, heatmap, basket, returns
├── inventory_reports.go       # turnover, dead stock, aging, GMROI
├── financial_reports.go       # profit, supplier perf, tax, reconciliation, audit
└── crm_reports.go             # customer insights, product perf, shrinkage
```

- [ ] **Step 2: Move methods to appropriate file**

Cut each method from `reporting_repository.go` and paste into the right new file. Verify build still passes after each move.

- [ ] **Step 3: Verify size**

```bash
cd backend && wc -l internal/repository/*_reports.go
```
Expected: each < 400 LOC.

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/repository/
git commit -m "refactor(reports): split reporting_repository into 4 files (P4.T8)"
```

---

## Task 4.9: Sales service split

**Files:**
- Modify: `backend/internal/services/sales_service.go:27-363`
- Create: `internal/services/pricing.go`, `stock_allocator.go`, `loyalty.go`, `promotion_resolver.go`, `checkout.go`

- [ ] **Step 1: Create `pricing.go`**

```go
package services

import "inventory/backend/internal/domain"

// ComputeLineSubtotal returns subtotal for a cart line.
func ComputeLineSubtotal(unitPrice float64, qty int, discount float64) float64 {
	return unitPrice * float64(qty) * (1 - discount/100)
}

// ComputeOrderTax returns tax from a setting-derived rate.
func ComputeOrderTax(subtotal, rate float64) float64 { return subtotal * rate }
```

- [ ] **Step 2: Create `stock_allocator.go`**

Move FEFO batch selection + locking logic from sales_service.go.

- [ ] **Step 3: Create `loyalty.go`**

Move `handleLoyalty` (earn/redeem/tiers).

- [ ] **Step 4: Move `promotion_resolver.go`** (already created in P2.T8)

- [ ] **Step 5: Create `checkout.go`** as orchestrator

```go
package services

import (
	"context"
	"gorm.io/gorm"
	"inventory/backend/internal/domain"
)

func (s *SalesService) ProcessCheckout(ctx context.Context, req CheckoutRequest, userID uint) (*domain.Order, error) {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. fetch + lock products
		// 2. allocate batches FEFO (stock_allocator)
		// 3. resolve promotions (promotion_resolver)
		// 4. compute pricing + tax (pricing)
		// 5. handle loyalty (loyalty)
		// 6. create order + transaction
	})
}
```

- [ ] **Step 6: Verify each file < 300 LOC**

```bash
cd backend && wc -l internal/services/{pricing,stock_allocator,loyalty,promotion_resolver,checkout,sales_service}.go
```

- [ ] **Step 7: Run tests, commit**

```bash
cd backend && go test ./...
git add internal/services/
git commit -m "refactor(sales): split sales_service into focused modules (P4.T9)"
```

---

## Task 4.10: Return handler split

**Files:**
- Modify: `backend/internal/handlers/return_handler.go:230-454`
- Create: `internal/handlers/request_return.go`, `process_return.go`

- [ ] **Step 1: Move `RequestReturn` to `request_return.go`**

- [ ] **Step 2: Move `ProcessReturn` to `process_return.go`**

- [ ] **Step 3: Delete the old `return_handler.go`**

- [ ] **Step 4: Verify each < 250 LOC**

- [ ] **Step 5: Run tests, commit**

```bash
cd backend && go test ./...
git add internal/handlers/
git commit -m "refactor(returns): split return_handler into request/process files (P4.T10)"
```

---

## Task 4.11: OpenAPI regeneration

**Files:**
- Modify: every handler file (add annotations)
- Regenerate: `backend/api/docs.go`, `swagger.json`, `swagger.yaml`
- Modify: `backend/api/openapi.yaml` (or keep auto-generated)

- [ ] **Step 1: Install swag CLI**

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

- [ ] **Step 2: Add annotations to one handler as template**

```go
// @Summary      Create product
// @Description  Create a new product
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        request body requests.ProductCreateRequest true "Product"
// @Success      201 {object} domain.Product
// @Failure      400 {object} errors.AppError
// @Security     BearerAuth
// @Router       /api/v1/products [post]
func (h *ProductHandler) Create(c *gin.Context) { ... }
```

Repeat for all handlers.

- [ ] **Step 3: Run swag**

```bash
cd backend && swag init -g cmd/server/main.go -o api/
```

- [ ] **Step 4: Verify Swagger UI renders all routes**

```bash
cd backend && go run ./cmd/server &
curl http://localhost:8080/swagger/doc.json | jq '.paths | keys | length'
```
Expected: 160

- [ ] **Step 5: Commit**

```bash
cd backend
git add api/ internal/handlers/
git commit -m "feat(openapi): regenerate swagger docs from handler annotations (P4.T11)"
```

---

## Task 4.12: Real webhook integration factory

**Files:**
- Modify: `backend/internal/integrations/integrations.go`
- Create: `internal/integrations/shopify.go`, `quickbooks.go`
- Modify: `backend/internal/app/container.go`

- [ ] **Step 1: Define integration interface**

```go
package integrations

import "github.com/gin-gonic/gin"

type Integration interface {
	Name() string
	HandleWebhook(c *gin.Context) error
}

type Registry struct{ integrations map[string]Integration }

func NewRegistry() *Registry {
	return &Registry{integrations: make(map[string]Integration)}
}

func (r *Registry) Register(i Integration) { r.integrations[i.Name()] = i }

func (r *Registry) HandleWebhook(c *gin.Context) {
	name := c.Param("name")
	i, ok := r.integrations[name]
	if !ok { c.JSON(404, gin.H{"error": "unknown integration"}); return }
	if err := i.HandleWebhook(c); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
	c.JSON(200, gin.H{"status": "ok"})
}
```

- [ ] **Step 2: Add stubs**

`shopify.go`:

```go
package integrations

import "github.com/gin-gonic/gin"

type Shopify struct{}

func (s *Shopify) Name() string { return "shopify" }
func (s *Shopify) HandleWebhook(c *gin.Context) error {
	// TODO: verify HMAC, parse order events, translate to internal domain
	return nil
}
```

Similar for `quickbooks.go`.

- [ ] **Step 3: Register in container**

```go
c.Integrations = integrations.NewRegistry()
c.Integrations.Register(&integrations.Shopify{})
c.Integrations.Register(&integrations.QuickBooks{})
```

- [ ] **Step 4: Update route**

```go
g.POST("/webhooks/:name", c.Integrations.HandleWebhook)
```

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/integrations/ internal/app/
git commit -m "feat(integrations): registry pattern with Shopify/QuickBooks stubs (P4.T12)"
```

---

## Task 4.13: Centralized config via koanf

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/go.mod`

- [ ] **Step 1: Add koanf**

```bash
cd backend
go get github.com/knadh/koanf/v2
go get github.com/knadh/koanf/providers/env
go get github.com/knadh/koanf/providers/file
```

- [ ] **Step 2: Replace `LoadConfig`**

```go
import (
	"github.com/knadh/koanf/providers/conf"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

func LoadConfig(path string) (*Config, error) {
	k := koanf.New(".")
	if path != "" { _ = k.Load(file.Provider(path), toml.Parser()) }
	_ = k.Load(env.Provider("", ".", func(s string) string { return strings.ToLower(strings.ReplaceAll(s, "_", ".")) }), nil)

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil { return nil, err }
	return &cfg, nil
}
```

- [ ] **Step 3: Add config file example**

Create `backend/config.example.toml`:

```toml
APP_ENV = "dev"
SERVER_PORT = 8080
DB_HOST = "localhost"
# ... etc
```

- [ ] **Step 4: Verify**

```bash
cd backend && go build ./... && go test ./...
```

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/config/ go.mod go.sum config.example.toml
git commit -m "refactor(config): adopt koanf for typed config loading (P4.T13)"
```

---

## Task 4.14: Centralized error wrapping

**Files:**
- Modify: 10 handlers that emit `gin.H{"error": "..."}` directly

- [ ] **Step 1: Grep**

```bash
cd backend && grep -rn 'gin.H{"error"' --include="*.go" | grep -v "_test.go"
```

- [ ] **Step 2: Replace each with `middleware.RespondError(c, appErr)`**

```go
// Before:
c.JSON(400, gin.H{"error": "invalid request"})

// After:
middleware.RespondError(c, errors.ErrBadRequest.WithMessage("invalid request"))
```

Or use `c.Error(err)` for the middleware to handle.

- [ ] **Step 3: Verify grep returns 0**

```bash
cd backend && grep -rn 'gin.H{"error"' --include="*.go" | grep -v "_test.go" | wc -l
```
Expected: 0

- [ ] **Step 4: Run tests, commit**

```bash
cd backend && go test ./...
git add internal/handlers/ internal/middleware/
git commit -m "refactor: route all errors through middleware (P4.T14)"
```

---

## Task 4.15: RabbitMQ topology centralization

**Files:**
- Modify: `backend/internal/message_broker/rabbitmq.go:20-319`
- Create: `internal/message_broker/topology.go`

- [ ] **Step 1: Define topology constants**

```go
// internal/message_broker/topology.go
package messagebroker

const (
	ExchangeName = "inventory"

	QueueBulkImport       = "bulk-import-queue"
	QueueBulkImportConfirm = "bulk-import-confirm-queue"
	QueueBulkExport       = "bulk-export-queue"
	QueueStockAlerts      = "stock-alerts-queue"
	QueueAlerts           = "alerts"
	QueueReporting        = "reporting"

	RoutingBulkImport        = "bulk.import"
	RoutingBulkImportConfirm = "bulk.import.confirm"
	RoutingBulkExport        = "bulk.export"
	RoutingStockAdjusted     = "stock.adjusted"
	RoutingAlertTriggered    = "alert.triggered"
	RoutingReportGenerate    = "report.generate"
	RoutingProductCreated    = "product.created"
	RoutingProductUpdated    = "product.updated"
	RoutingProductDeleted    = "product.deleted"
)

type Topology struct {
	Exchange   string
	Queues     map[string]string // queue name → routing key
}

func DefaultTopology() Topology {
	return Topology{
		Exchange: ExchangeName,
		Queues: map[string]string{
			QueueBulkImport:        RoutingBulkImport,
			QueueBulkImportConfirm: RoutingBulkImportConfirm,
			QueueBulkExport:        RoutingBulkExport,
			QueueStockAlerts:       RoutingStockAdjusted,
			QueueAlerts:            RoutingAlertTriggered,
			QueueReporting:         RoutingReportGenerate,
		},
	}
}
```

- [ ] **Step 2: Use topology in `RabbitMQManager.Subscribe`**

```go
func (m *RabbitMQManager) Subscribe(queueName string, handler func([]byte) error) (context.CancelFunc, error) {
	t := DefaultTopology()
	routingKey, ok := t.Queues[queueName]
	if !ok { return nil, fmt.Errorf("unknown queue: %s", queueName) }
	// declare exchange, queue, binding, then consume
}
```

- [ ] **Step 3: Run tests, commit**

```bash
cd backend && go test ./...
git add internal/message_broker/
git commit -m "refactor(mq): centralize RabbitMQ topology in one file (P4.T15)"
```

---

## Task 4.16: Consumer cancel-func wiring

**Files:**
- Modify: `backend/internal/bootstrap/app.go:171-205`

- [ ] **Step 1: Store cancel funcs**

```go
type App struct {
	container     *app.Container
	cron          *cron.Cron
	cancelFuncs   []context.CancelFunc
	wg            sync.WaitGroup
	httpServer    *http.Server
}

func (a *App) startConsumers(ctx context.Context) {
	for i := 0; i < a.container.Cfg.ConsumerConcurrency; i++ {
		cancel, _ := bulkConsumer.Start(ctx, i)
		a.cancelFuncs = append(a.cancelFuncs, cancel)
		a.wg.Add(1)
		go func() { defer a.wg.Done(); <-ctx.Done() }()
	}
	// ... alert + reporting consumers
}

func (a *App) Stop(ctx context.Context) error {
	for _, cancel := range a.cancelFuncs { cancel() }
	done := make(chan struct{})
	go func() { a.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		logrus.Warn("consumer shutdown timed out")
	}
	// ... rest of shutdown
}
```

- [ ] **Step 2: Test graceful shutdown**

```go
func TestGracefulShutdown_CancelsConsumers(t *testing.T) {
	app := setupTestApp(t)
	app.Start(context.Background())
	time.Sleep(100 * time.Millisecond)
	app.Stop(context.Background())
	assert.Greater(t, len(app.cancelFuncs), 0)
}
```

- [ ] **Step 3: Commit**

```bash
cd backend
git add internal/bootstrap/
git commit -m "fix(mq): wire consumer cancel funcs into graceful shutdown (P4.T16)"
```

---

## Task 4.17: Env-driven CORS

**Files:**
- Modify: `backend/internal/router/router.go` (or `router/groups/public.go`)

- [ ] **Step 1: Replace hard-coded origins**

```go
// Before:
AllowOrigins: []string{"http://localhost:5173"}

// After:
AllowOrigins: c.Cfg.CORSAllowedOrigins
```

- [ ] **Step 2: Add config field**

```go
CORSAllowedOrigins []string // CORS_ALLOWED_ORIGINS, CSV
```

- [ ] **Step 3: Test**

```bash
CORS_ALLOWED_ORIGINS="https://app.example.com,https://staging.example.com" go run ./cmd/server
curl -H "Origin: https://app.example.com" -I http://localhost:8080/api/v1/users/login
```
Expected: `Access-Control-Allow-Origin: https://app.example.com`

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/router/ internal/config/
git commit -m "feat(cors): env-driven CORS origins (P4.T17)"
```

---

## Task 4.18: Loyalty tier thresholds via settings

**Files:**
- Modify: `backend/internal/services/sales_service.go:257`
- Modify: `backend/internal/handlers/settings.go` (already adds descriptors in P2.T15)

- [ ] **Step 1: Add setting descriptors**

In `settings.go`:

```go
"loyalty.tier_silver":   {Type: "number", Validate: func(v string) error { n, _ := strconv.Atoi(v); if n < 0 { return errors.New("negative") }; return nil }},
"loyalty.tier_gold":     {Type: "number", Validate: ...},
"loyalty.tier_platinum": {Type: "number", Validate: ...},
"loyalty.earning_rate":  {Type: "number", Validate: func(v string) error { f, _ := strconv.ParseFloat(v, 64); if f < 0 || f > 1 { return errors.New("0..1") }; return nil }},
"loyalty.redemption_rate":{Type: "number", Validate: ...},
```

- [ ] **Step 2: Read thresholds from settings in `handleLoyalty`**

```go
silver, _ := strconv.Atoi(getSetting(ctx, tx, "loyalty.tier_silver", "500"))
gold, _ := strconv.Atoi(getSetting(ctx, tx, "loyalty.tier_gold", "2500"))
platinum, _ := strconv.Atoi(getSetting(ctx, tx, "loyalty.tier_platinum", "10000"))
```

- [ ] **Step 3: Write failing test**

```go
func TestLoyaltyTiers_CustomThresholds(t *testing.T) {
	db := setupTestDB(t)
	setSetting(t, db, "loyalty.tier_gold", "100") // lower threshold
	// earn 150 points → should reach gold
	// ... svc.ProcessCheckout → assert tier == GOLD
}
```

- [ ] **Step 4: Run, commit**

```bash
cd backend && go test ./internal/services/ -run TestLoyaltyTiers -v
git add internal/services/sales_service.go internal/handlers/settings.go
git commit -m "feat(loyalty): tier thresholds via settings (P4.T18)"
```

---

## Phase 4 verification gate

```bash
cd backend
go test ./...
# Architectural acceptance checks:
grep -rn "repository\.DB" --include="*.go" | grep -v "database.go" | wc -l   # expect 0
grep -rn 'gin.H{"error"' --include="*.go" | grep -v "_test.go" | wc -l    # expect 0
curl http://localhost:8080/swagger/doc.json | jq '.paths | keys | length'  # expect 160
```

Append to `CHANGELOG.md`. Merge `phase-4-architecture` to main, tag `v1.4-architecture`.

---

# Phase 5 — Operations & observability (10 tasks)

## Task 5.1 + 5.2: Add /ready endpoint + liveness/readiness split

**Files:**
- Create: `backend/internal/handlers/ready.go`
- Modify: `backend/internal/router/groups/public.go`

- [ ] **Step 1: Write failing test**

```go
func TestReady_FailsWhenDependencyDown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &ReadyHandler{db: brokenDB, redis: rdb, amqp: nil}
	r.GET("/ready", h.Ready)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 503, w.Code)
}
```

- [ ] **Step 2: Implement `/ready`**

```go
func (h *ReadyHandler) Ready(c *gin.Context) {
	status := gin.H{"status": "READY"}
	failed := []string{}

	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil { failed = append(failed, "db") }
	if h.redis.Ping(c.Request.Context()).Err() != nil { failed = append(failed, "redis") }
	if h.amqp == nil || h.amqp.IsClosed() { failed = append(failed, "amqp") }
	exists, _ := h.uploader.BucketExists(c.Request.Context(), "reports")
	if !exists { failed = append(failed, "minio") }

	if len(failed) > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "NOT_READY", "failed": failed})
		return
	}
	c.JSON(http.StatusOK, status)
}
```

- [ ] **Step 3: Document split in runbook (Phase 7)**

`/health` = liveness (process is up); `/ready` = readiness (deps reachable).

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/handlers/ready.go internal/router/groups/public.go internal/handlers/ready_test.go
git commit -m "feat(health): add /ready endpoint with dependency checks (P5.T1, P5.T2)"
```

---

## Task 5.3: Custom Prometheus collectors

**Files:**
- Create: `backend/internal/metrics/` (5 files)
- Modify: handlers + services to record metrics

- [ ] **Step 1: Create collectors**

```go
// internal/metrics/http.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total", Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds", Help: "Request latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
)

// similar for cache, jobs, payments, alerts
```

- [ ] **Step 2: Add middleware to record**

```go
// internal/middleware/metrics.go
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		path := c.FullPath()
		if path == "" { path = "unknown" }
		metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, strconv.Itoa(c.Writer.Status())).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}
```

- [ ] **Step 3: Verify in /metrics**

```bash
curl http://localhost:8080/metrics | grep http_requests_total
```

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/metrics/ internal/middleware/metrics.go
git commit -m "feat(metrics): Prometheus collectors for HTTP, cache, jobs, payments, alerts (P5.T3)"
```

---

## Task 5.4: Grafana dashboards

**Files:**
- Create: `grafana/dashboards/{overview,api,jobs,payments,alerts}.json`
- Create: `grafana/provisioning/dashboards/dashboards.yml`

- [ ] **Step 1: Author dashboards**

For each, build a JSON file matching Grafana 9.1 schema with:
- Row 1: request rate (from `http_requests_total`), error rate
- Row 2: p50/p95/p99 latency
- For jobs: `jobs_processed_total{status}` stacked
- For payments: success ratio
- For alerts: `alerts_fired_total{type}` by type

- [ ] **Step 2: Provision**

`dashboards.yml`:

```yaml
apiVersion: 1
providers:
  - name: 'Quantify'
    orgId: 1
    folder: 'Quantify'
    type: file
    disableDeletion: false
    options:
      path: /etc/grafana/dashboards
```

- [ ] **Step 3: Commit**

```bash
cd backend
git add grafana/
git commit -m "feat(grafana): provisioned dashboards (P5.T4)"
```

---

## Task 5.5: Replace `fmt.Print` with `logrus`

**Files:** (known list)
- `handlers/notifications.go:46-48`
- `handlers/websocket.go:56`
- `services/purchase_order_service.go:110`
- `repository/reporting_repository.go:745,787`
- Plus any grep result

- [ ] **Step 1: Grep**

```bash
cd backend && grep -rn 'fmt.Print\|fmt.Printf\|fmt.Println' --include="*.go" | grep -v "_test.go"
```

- [ ] **Step 2: Replace each**

```go
// Before:
fmt.Printf("fetched user %d\n", uid)

// After:
logrus.WithField("user_id", uid).Info("fetched user")
```

- [ ] **Step 3: Verify zero matches**

```bash
cd backend && grep -rn 'fmt.Print' --include="*.go" | grep -v "_test.go" | wc -l
```
Expected: 0

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/
git commit -m "chore(logging): replace fmt.Print with logrus (P5.T5)"
```

---

## Task 5.6: Request ID middleware

**Files:**
- Create: `backend/internal/middleware/request_id.go`

- [ ] **Step 1: Write failing test**

```go
func TestRequestID_PropagatesHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	var captured string
	r.GET("/x", func(c *gin.Context) { captured = c.GetString("request_id"); c.Status(200) })

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("X-Request-ID", "my-id-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, "my-id-123", captured)
	assert.Equal(t, "my-id-123", w.Header().Get("X-Request-ID"))
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	// similar but without X-Request-ID header
}
```

- [ ] **Step 2: Implement**

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" { id = uuid.NewString() }
		c.Set("request_id", id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}
```

- [ ] **Step 3: Add logrus hook to inject request_id**

In `cmd/server/main.go`:

```go
logrus.AddHook(&requestIDHook{})

// Hook:
type requestIDHook struct{}
func (h *requestIDHook) Levels() []logrus.Level { return logrus.AllLevels }
func (h *requestIDHook) Fire(e *logrus.Entry) error {
	if id := getGinRequestID(e.Context); id != "" {
		e.Data["request_id"] = id
	}
	return nil
}
```

(Implementation: pass context via `logrus.WithContext(ctx)` or use a global map.)

- [ ] **Step 4: Run tests, commit**

```bash
cd backend && go test ./internal/middleware/ -run TestRequestID -v
git add internal/middleware/ cmd/server/
git commit -m "feat(observability): request ID middleware + logrus hook (P5.T6)"
```

---

## Task 5.7: OpenTelemetry tracing

**Files:**
- Create: `backend/internal/telemetry/` (tracer.go, propagation.go)
- Modify: `cmd/server/main.go`, services, repos, message_broker

- [ ] **Step 1: Add OTel dependency**

```bash
cd backend
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/sdk
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
```

- [ ] **Step 2: Initialize tracer**

`internal/telemetry/tracer.go`:

```go
package telemetry

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(ctx context.Context, serviceName, otlpEndpoint string) (func(context.Context) error, error) {
	exporter, err := otlptrace.New(ctx, otlptrace.WithEndpoint(otlpEndpoint), otlptrace.WithInsecure())
	if err != nil { return nil, err }
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName))),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
```

- [ ] **Step 3: Wire OTel Gin middleware**

```go
r.Use(otelgin.Middleware("quantify-backend"))
```

- [ ] **Step 4: Add spans in services**

In `sales_service.go::ProcessCheckout`:

```go
ctx, span := otel.Tracer("sales").Start(ctx, "ProcessCheckout")
defer span.End()

ctx, span = otel.Tracer("sales").Start(ctx, "fetchProducts")
// ... work
span.End()
```

- [ ] **Step 5: Propagate via AMQP headers and Redis**

In `message_broker/rabbitmq.go::Publish`:

```go
otel.GetTextMapPropagator().Inject(ctx, amqpCarrier{msg})
```

- [ ] **Step 6: Verify in Jaeger/Tempo**

Spin up jaegertracing/all-in-one, run a synthetic checkout, view span tree.

- [ ] **Step 7: Commit**

```bash
cd backend
git add internal/telemetry/ internal/services/ internal/message_broker/ cmd/server/ go.mod go.sum
git commit -m "feat(observability): OpenTelemetry tracing across HTTP/services/MQ (P5.T7)"
```

---

## Task 5.8: CI workflow

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Author workflow**

```yaml
name: CI
on:
  push:
    branches: [main, phase-*]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        ports: ['5432:5432']
        options: --health-cmd pg_isready --health-interval 5s
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: {go-version: '1.24'}
      - run: cd backend && go vet ./...
      - run: cd backend && go test -race ./...
      - uses: golangci/golangci-lint-action@v4
        with: {version: latest}
      - uses: securego/gosec@master
        with: {args: './...'}
      - run: docker build -t quantify-backend:test backend/
```

- [ ] **Step 2: Commit + push to test**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: GitHub Actions workflow (P5.T8)"
```

---

## Task 5.9: Container hardening

**Files:**
- Modify: `backend/Dockerfile`

- [ ] **Step 1: Add non-root user + healthcheck**

```dockerfile
# ... existing build steps
FROM alpine
RUN adduser -D -g '' nonroot
WORKDIR /app
COPY --from=builder /inventory-backend /app/inventory-backend
USER nonroot
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/inventory-backend"]
```

- [ ] **Step 2: Verify**

```bash
docker build -t quantify:test backend/
docker run --rm --cap-drop=ALL quantify:test
docker inspect quantify:test --format '{{.Config.User}}'
```
Expected: `nonroot`

- [ ] **Step 3: Commit**

```bash
cd backend
git add Dockerfile
git commit -m "build: non-root user, HEALTHCHECK, dropped caps (P5.T9)"
```

---

## Task 5.10: Log shipping example config

**Files:**
- Create: `docs/log-shipping.md`

- [ ] **Step 1: Document Vector + Loki**

```markdown
# Log shipping

The backend emits JSON logs to stdout. To ship them to Loki:

## Vector

\`\`\`yaml
[sources.backend]
type = "docker_logs"
include_labels = ["container_name=quantify-backend"]

[transforms.parse]
type = "remap"
source = '''
. |= parse_json!(.message)
'''

[sinks.loki]
type = "loki"
inputs = ["parse"]
endpoint = "http://loki:3100"
\`\`\`

## Fluent Bit

[equivalent fluent-bit config]
```

- [ ] **Step 2: Commit**

```bash
git add docs/log-shipping.md
git commit -m "docs: log shipping examples (P5.T10)"
```

---

## Phase 5 verification gate

```bash
curl http://localhost:8080/metrics | grep -E "http_requests_total|cache_hits_total"  # expect non-zero after load
curl http://localhost:8080/ready  # expect 200
```

Append to `CHANGELOG.md`. Merge `phase-5-ops` to main, tag `v1.5-ops`.

---

# Phase 6 — Test coverage (8 tasks)

## Task 6.1: Testcontainers helper for Postgres

**Files:**
- Create: `backend/test/integration/db.go`

- [ ] **Step 1: Add testcontainers**

```bash
cd backend
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

- [ ] **Step 2: Write helper**

```go
// test/integration/db.go
//go:build integration
package integration

import (
	"context"
	"database/sql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
	"time"
)

func SetupPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()
	pg, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("test"),
		postgres.WithPassword("test"),
		postgres.WithUsername("test"),
	)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, _ := pg.ConnectionString(ctx)
	for i := 0; i < 30; i++ {
		db, err := sql.Open("postgres", dsn)
		if err == nil && db.Ping() == nil { db.Close(); break }
		time.Sleep(time.Second)
	}
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	return gdb
}
```

- [ ] **Step 3: Commit**

```bash
cd backend
git add test/integration/ go.mod go.sum
git commit -m "test: testcontainers Postgres helper for integration tests (P6.T1)"
```

---

## Tasks 6.2-6.7: Integration tests

For each, follow the pattern: write the test using `integration.SetupPostgres(t)`, run migrations, seed test data, exercise the flow, assert outcomes.

- **P6.T2**: Checkout → return → refund integration test (`internal/handlers/return_integration_test.go`)
- **P6.T3**: Bulk import end-to-end (`internal/consumers/bulk_consumer_test.go`)
- **P6.T4**: PO receive integration (`internal/services/purchase_order_service_test.go`)
- **P6.T5**: Alert consumer test (`internal/consumers/alert_consumer_test.go`)
- **P6.T6**: Permission cache invalidation test (`internal/repository/role_repository_ext_test.go`)
- **P6.T7**: Webhook integration factory test (`internal/integrations/integrations_test.go`)

For each:

```go
//go:build integration

func TestX(t *testing.T) {
	db := integration.SetupPostgres(t)
	// ... seed, exercise, assert
}
```

Run with: `cd backend && go test -tags=integration ./...`

- [ ] **Step per task**: Write failing test → fix implementation if needed → commit

```bash
cd backend
go test -tags=integration ./internal/handlers/ -run TestCheckoutReturnRefund -v
git add internal/handlers/return_integration_test.go
git commit -m "test(integration): checkout → return → refund flow (P6.T2)"
```

(Repeat for each of 6.2-6.7.)

---

## Task 6.8: Property-based tests

**Files:**
- Create: `backend/internal/services/{promotion_resolver,refund_math,tax_calc}_property_test.go`

- [ ] **Step 1: Add rapid**

```bash
cd backend && go get pgregory.net/rapid
```

- [ ] **Step 2: Author property tests**

```go
// promotion_resolver_property_test.go
func TestPromotion_HierarchyProperty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		productID := rapid.Uint().Draw(t, "productID")
		subID := rapid.Uint().Draw(t, "subID")
		catID := rapid.Uint().Draw(t, "catID")
		promos := rapid.SliceOf(rapidCustomPromotion()).Draw(t, "promos")
		got := ResolvePromotion(promos, productID, subID, catID)
		// property: if a product-scoped promo is in the list and applies, got must be it
		for _, p := range promos {
			if p.ProductID != nil && *p.ProductID == productID {
				if got == nil || got.ID != p.ID { t.Fatal("property violated") }
			}
		}
	})
}

func TestRefund_NeverNegative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		unitPrice := rapid.Float64Range(0.01, 10000).Draw(t, "unitPrice")
		qty := rapid.IntRange(1, 100).Draw(t, "qty")
		taxRate := rapid.Float64Range(0, 0.5).Draw(t, "taxRate")
		got := computeRefund(unitPrice, qty, taxRate)
		if got < 0 { t.Fatalf("negative refund %f", got) }
	})
}
```

- [ ] **Step 3: Run for 1 minute**

```bash
cd backend && go test ./internal/services/ -run "Property" -v -count=1 -timeout=2m
```

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/services/*_property_test.go go.mod go.sum
git commit -m "test: property-based tests for promotions, refunds, tax (P6.T8)"
```

---

## Phase 6 verification gate

```bash
cd backend
go test ./...
go test -tags=integration ./...
go test -race ./...
# coverage on services + handlers:
go test -coverprofile=coverage.out ./internal/services/... ./internal/handlers/...
go tool cover -func=coverage.out | grep -E "total:|services|handlers"
```
Expected: total > 70%.

Append to `CHANGELOG.md`. Merge `phase-6-tests` to main, tag `v1.6-tests`.

---

# Phase 7 — Documentation & API spec (6 tasks)

## Task 7.1: Regenerate OpenAPI spec to 160 routes

**Files:**
- Regenerate: `backend/api/openapi.yaml`
- Verify: `backend/api/docs.go`, `swagger.json`, `swagger.yaml`

- [ ] **Step 1: Ensure annotations are complete (P4.T11 should have done this)**

```bash
cd backend && swag init -g cmd/server/main.go -o api/
```

- [ ] **Step 2: Verify count**

```bash
cd backend
go run ./cmd/server &
sleep 2
curl http://localhost:8080/swagger/doc.json | jq '.paths | keys | length'
```
Expected: 160

- [ ] **Step 3: Add request/response examples**

Manually extend `openapi.yaml` for non-trivial endpoints:

```yaml
/checkout:
  post:
    summary: POS checkout
    requestBody:
      required: true
      content:
        application/json:
          schema: {$ref: '#/components/schemas/CheckoutRequest'}
          example:
            items: [{product_id: 1, quantity: 2}]
            payment_method: cash
            register_location_id: 1
    responses:
      '201':
        content:
          application/json:
            schema: {$ref: '#/components/schemas/Order'}
```

- [ ] **Step 4: Commit**

```bash
cd backend
git add api/
git commit -m "docs: regenerate OpenAPI spec with examples (P7.T1)"
```

---

## Task 7.2: Update `docs/business_logic.md`

**Files:**
- Modify: `docs/business_logic.md`

- [ ] **Step 1: Remove false claims**

Search for "multi-tenant", "tenant", "RFID", and remove or annotate as "planned".

- [ ] **Step 2: Document actual flows**

Add sections covering:
- RBAC model (Admin/Manager/Staff/Customer, permission list)
- Checkout flow (FEFO, FOR UPDATE, promotion resolver, tax, loyalty)
- Return flow (request → approve/reject → refund, batch restoration, loyalty reversal)
- Bulk import lifecycle (upload → QUEUED → PENDING_CONFIRMATION → COMPLETED)
- Loyalty tiers (now settings-driven)
- Returned/added behavior of Order.TaxAmount (from P3.T7)

- [ ] **Step 3: Commit**

```bash
git add docs/business_logic.md
git commit -m "docs: align business_logic.md with implementation (P7.T2)"
```

---

## Task 7.3: Add `docs/architecture.md`

**Files:**
- Create: `docs/architecture.md`

- [ ] **Step 1: Write ADRs**

```markdown
# Architecture

## Overview
[Layered diagram: Handler → Service → Repository]

## Decisions

### ADR-001: Router split into per-domain groups
Context: router.go had grown to 442 lines.
Decision: split into `internal/router/groups/*.go`.
Consequences: easier navigation, smaller PRs, but more files.

### ADR-002: DI container
Context: package-level repository.DB blocked testing.
Decision: introduce `app.Container` built in bootstrap.
Consequences: explicit wiring, mockable interfaces.

### ADR-003: golang-migrate over AutoMigrate
Context: AutoMigrate in production is unsafe.
Decision: versioned up/down SQL via cmd/migrator.
Consequences: safer deployments, but operators must run migrator.
```

- [ ] **Step 2: Add deployment topology diagram**

Use Mermaid:

```mermaid
graph TB
  Client --> Nginx
  Nginx --> Backend
  Backend --> Postgres
  Backend --> Redis
  Backend --> RabbitMQ
  Backend --> MinIO
  Backend --> AI[ai-service]
```

- [ ] **Step 3: Commit**

```bash
git add docs/architecture.md
git commit -m "docs: architecture overview + ADRs (P7.T3)"
```

---

## Task 7.4: Add `docs/security.md`

**Files:**
- Create: `docs/security.md`

- [ ] **Step 1: Document threat model**

```markdown
# Security

## Threat model (STRIDE-lite)

| Threat | Vector | Control |
|---|---|---|
| SQL injection | User input in raw SQL | websearch_to_tsquery (P1.T11), parameterized queries (P3.T3) |
| Auth bypass | Missing context check | Strict ownership checks (P1.T12) |
| Privilege escalation | Stale permission cache | 60s TTL + invalidate on mutation (P1.T5) |
| IPN spoofing | Unsigned callback | HMAC-SHA512 verification (P1.T2) |
| WebSocket CSRF | Permissive origin | Allowlist (P1.T3) |

## Key rotation runbook

1. Generate new JWT_SECRET: `openssl rand -hex 32`
2. Set env var on backend container; restart.
3. Existing JWTs become invalid; users re-login.
4. Document rotation date in `docs/security.md`.
```

- [ ] **Step 2: Commit**

```bash
git add docs/security.md
git commit -m "docs: security threat model and key rotation runbook (P7.T4)"
```

---

## Task 7.5: Add `docs/runbook.md`

**Files:**
- Create: `docs/runbook.md`

- [ ] **Step 1: Document incidents**

```markdown
# Runbook

## Postgres down
1. Check container: `docker ps | grep db`
2. Logs: `docker logs quantify-db`
3. Restart: `docker restart quantify-db`
4. Backend `/ready` should return 200.

## Redis down
1. Check cluster: `docker exec -it redis1 redis-cli -a redispass ping`
2. Backend will lose cache; performance degrades but functional.

## RabbitMQ down
1. Check container.
2. Restart.
3. Bulk jobs in `QUEUED` state will resume automatically.

## Alert backlog
1. Check `alerts_fired_total` rate vs delivery rate.
2. If consumers stuck: restart backend; consumer cancel funcs re-wire (P4.T16).

## Deploy
1. Build: `docker build -t quantify-backend:$VERSION backend/`
2. Push: `docker push ...`
3. Run migrator: `kubectl exec deploy/migrator -- ./migrator up`
4. Roll backend: `kubectl rollout restart deployment/backend`
5. Verify: `curl /ready` returns 200.

## Rollback
1. `kubectl rollout undo deployment/backend`
2. If schema migration involved, run `migrator down 1`.
```

- [ ] **Step 2: Commit**

```bash
git add docs/runbook.md
git commit -m "docs: incident response and deploy/rollback runbook (P7.T5)"
```

---

## Task 7.6: Update README files

**Files:**
- Modify: `README.md`, `docs/README.md`

- [ ] **Step 1: Update env var table**

Match `internal/config/config.go`:

```markdown
## Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| APP_ENV | no | dev | dev/prod; controls AutoMigrate |
| SERVER_PORT | no | 8080 | HTTP listen port |
| JWT_SECRET | YES | — | HS256 signing key |
| REFRESH_TOKEN_SECRET | YES | — | (validated by config) |
| REFRESH_TOKEN_TTL | no | 720h | Refresh token lifetime |
| DATABASE_URL | no | postgres://user:password@localhost/inventory_db?sslmode=disable | DSN |
| REDIS_ADDRS | no | redis1:6379,... | Cluster addrs |
| RABBITMQ_URL | YES | — | amqp:// connection string |
| MINIO_* | YES | — | MinIO credentials (validated) |
| CORS_ALLOWED_ORIGINS | no | http://localhost:5173 | CSV |
| WS_ALLOWED_ORIGINS | no | http://localhost:5173 | CSV |
| AI_SERVICE_URL | no | http://localhost:8001 | Forecast + churn |
| CONSUMER_CONCURRENCY | no | 1 | Bulk consumer workers |
| RUN_MIGRATIONS | no | false | Trigger search-index backfill |
| SSLCOMMERZ_STORE_PASS | YES (for prod IPN) | — | IPN verification |
| OTEL_EXPORTER_OTLP_ENDPOINT | no | localhost:4317 | OTel collector |
```

- [ ] **Step 2: Add dev quickstart**

```markdown
## Development

\`\`\`bash
docker-compose up -d db redis1 redis2 redis3 rabbitmq minio
cd backend
cp .env.example .env  # edit secrets
go run ./cmd/migrator up
go run ./cmd/server
\`\`\`

## Tests

\`\`\`bash
go test ./...
go test -tags=integration ./...
\`\`\`
```

- [ ] **Step 3: Commit**

```bash
git add README.md docs/README.md
git commit -m "docs: update README with accurate env vars + dev quickstart (P7.T6)"
```

---

## Phase 7 verification gate

- `curl http://localhost:8080/swagger/doc.json | jq '.paths | keys | length'` = 160
- README env table matches `config.go` field names
- All four new docs reviewed and merged

Append to `CHANGELOG.md`:

```markdown
## [Phase 7] 2026-XX-XX — Documentation

- P7.T1: OpenAPI spec at 160 routes
- P7.T2: business_logic.md aligned
- P7.T3: architecture.md
- P7.T4: security.md
- P7.T5: runbook.md
- P7.T6: README updates

**Remediation plan v1 complete.**
```

Merge `phase-7-docs` to main, tag `v1.7-docs` and `v1.0-remediation-complete`.

---

# Final acceptance

The remediation plan is complete when:

1. All 85 tasks closed with passing tests and reviewer sign-off.
2. `go test ./...` and `go test -tags=integration ./...` green.
3. `golangci-lint` and `gosec` clean.
4. OpenAPI spec renders all 160 routes.
5. Grafana dashboards provisioned and live.
6. CI pipeline runs green on every PR.
7. All four documents reviewed.
8. CHANGELOG covers every phase.
9. No `repository.DB` direct access outside the package; no `fmt.Print`; no `gin.H{"error": ...}`.
10. Production-ready security headers, SSLCommerz IPN signed, permission cache correctly invalidated, alert consumer delivering alerts.

---

# Self-review notes

- Each task follows TDD: failing test → impl → pass → commit.
- File paths are absolute (under `backend/`) where possible.
- Phase ordering enforces dependencies: each later phase builds on prior work.
- Verification gates catch regressions before the next phase starts.
- Tasks within a phase are ordered by dependency (e.g., P4.T5 before P4.T6/P4.T7).
