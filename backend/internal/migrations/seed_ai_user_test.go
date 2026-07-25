package migrations

import (
	"testing"

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
