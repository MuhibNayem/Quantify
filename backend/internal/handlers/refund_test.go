package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeRefund_AdditiveTaxOnSubtotal(t *testing.T) {
	// unitPrice=99.99, qty=1, tax=7.5%
	// subtotal = 99.99; tax_share = 99.99 * 0.075 = 7.49925
	// total = 99.99 + 7.49925 = 107.48925 → round to 107.49
	got := computeRefund(99.99, 1, 0.075)
	assert.InDelta(t, 107.49, got, 0.01)
}

func TestComputeRefund_NoTax(t *testing.T) {
	got := computeRefund(100, 2, 0)
	assert.Equal(t, 200.0, got)
}

func TestComputeRefund_RoundsToCents(t *testing.T) {
	// Avoid floating-point drift on common values.
	got := computeRefund(10.00, 3, 0.0825)
	assert.InDelta(t, 32.48, got, 0.01) // 30 + 2.475
}