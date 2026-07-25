package migrations

import (
	"testing"

	"inventory/backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestSeedAIUser_RandomPasswordIsUnique(t *testing.T) {
	p1 := GenerateRandomPassword()
	p2 := GenerateRandomPassword()
	assert.NotEqual(t, p1, p2, "two generations must differ")
	assert.GreaterOrEqual(t, len(p1), 24, "generated password must be >=24 chars")
}

func TestSeedAIUser_HashIsBcrypt(t *testing.T) {
	password := GenerateRandomPassword()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	assert.NoError(t, err)
}

func TestSeedAIUser_DefaultInactive(t *testing.T) {
	user := buildAIUser("pwhash", domain.Role{})
	assert.False(t, user.IsActive, "AI agent user must default to inactive")
}
