#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 <base-url>"
  echo "  base-url: e.g. http://localhost:8080 or https://api.quantify.com"
  exit 1
}

[ $# -eq 1 ] || usage

BASE_URL="${1%/}"
PASS=0
FAIL=0
TOKEN=""

check() {
  local name="$1"
  local method="$2"
  local path="$3"
  local expected_code="$4"
  local data="${5:-}"

  local args=(-s -o /dev/null -w "%{http_code}" -X "$method")
  if [ -n "$TOKEN" ]; then
    args+=(-H "Authorization: Bearer $TOKEN")
  fi
  if [ -n "$data" ]; then
    args+=(-H "Content-Type: application/json" -d "$data")
  fi

  local code
  code=$(curl "${args[@]}" "${BASE_URL}${path}" 2>/dev/null || echo "000")

  if [ "$code" = "$expected_code" ]; then
    echo "  PASS  ${name} (${code})"
    PASS=$((PASS + 1))
  else
    echo "  FAIL  ${name} (expected ${expected_code}, got ${code})"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== Smoke Tests: ${BASE_URL} ==="

check "GET /health" GET "/health" "200"
check "GET /ready" GET "/ready" "200"
check "GET /metrics" GET "/metrics" "200"

echo ""
echo "--- Auth flow ---"

LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"smoke-test@quantify.local","password":"smoke-test-pass"}' 2>/dev/null || echo "{}")

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // .access_token // empty' 2>/dev/null || echo "")

if [ -n "$TOKEN" ]; then
  echo "  PASS  Login flow (token acquired)"
  PASS=$((PASS + 1))
else
  echo "  SKIP  Login flow (no test credentials — skipping authenticated tests)"
fi

if [ -n "$TOKEN" ]; then
  echo ""
  echo "--- Product CRUD ---"

  CREATE_RESPONSE=$(curl -s -X POST "${BASE_URL}/products" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"Smoke Test Product","sku":"SMOKE-'$$'","price":9.99,"quantity":10}' 2>/dev/null || echo "{}")

  PRODUCT_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id // .data.id // empty' 2>/dev/null || echo "")

  if [ -n "$PRODUCT_ID" ]; then
    echo "  PASS  Create product (${PRODUCT_ID})"
    PASS=$((PASS + 1))

    check "GET product" GET "/products/${PRODUCT_ID}" "200"

    check "Update product" PUT "/products/${PRODUCT_ID}" "200" \
      '{"name":"Smoke Test Product Updated","price":19.99}'

    check "Delete product" DELETE "/products/${PRODUCT_ID}" "200"
  else
    echo "  FAIL  Create product (no ID returned)"
    FAIL=$((FAIL + 1))
  fi

  echo ""
  echo "--- Checkout flow ---"

  check "List products" GET "/products" "200"
  check "Create checkout" POST "/checkout" "201" \
    '{"items":[{"product_id":"1","quantity":1}]}'
fi

echo ""
echo "=== Results: ${PASS} passed, ${FAIL} failed ==="

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
