package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserAfterFind_EmptyRoleWhenNotPreloaded(t *testing.T) {
	u := User{Username: "alice"}
	_ = u.AfterFind(nil)
	assert.Equal(t, "", u.RoleName)
}

func TestUserAfterFind_PopulatesRoleNameWhenPreloaded(t *testing.T) {
	u := User{
		Username: "alice",
		Role:     Role{Model: gorm.Model{ID: 1}, Name: "Admin"},
	}
	_ = u.AfterFind(nil)
	assert.Equal(t, "Admin", u.RoleName)
}

func TestUserAfterFind_ClearsStaleRoleNameWhenNotPreloaded(t *testing.T) {
	u := User{Username: "alice", RoleName: "stale-value"}
	_ = u.AfterFind(nil)
	assert.Equal(t, "", u.RoleName)
}
