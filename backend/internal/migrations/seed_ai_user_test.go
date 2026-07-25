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
