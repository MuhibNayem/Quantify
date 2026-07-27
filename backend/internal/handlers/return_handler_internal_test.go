package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeLoyaltyReversal_Proportional(t *testing.T) {
	assert.Equal(t, 40, computeLoyaltyReversal(100, 10, 4))
	assert.Equal(t, 100, computeLoyaltyReversal(100, 10, 10))
	assert.Equal(t, 20, computeLoyaltyReversal(80, 4, 1))
}

func TestComputeLoyaltyReversal_ZeroPointsRedeemed(t *testing.T) {
	assert.Equal(t, 0, computeLoyaltyReversal(0, 10, 1))
}

func TestComputeLoyaltyReversal_ZeroQuantity(t *testing.T) {
	assert.Equal(t, 0, computeLoyaltyReversal(100, 0, 1))
}

func TestComputeLoyaltyReversal_FullReturnOfPartialRedemption(t *testing.T) {
	assert.Equal(t, 50, computeLoyaltyReversal(50, 10, 10))
}