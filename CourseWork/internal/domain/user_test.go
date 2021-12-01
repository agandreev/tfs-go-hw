package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var user *User

func setupUser() {
	user = NewUser("username")
}

func TestNewUser(t *testing.T) {
	newUser := NewUser("new user")
	assert.NotNil(t, newUser.limits)
}

func TestUser_AddLimit(t *testing.T) {
	setupUser()
	eps := float64(7.)/3 - float64(4.)/3 - 1.
	err := user.AddLimit("first pair", 0-eps)
	assert.Error(t, err)
	err = user.AddLimit("first pair", 1+eps)
	assert.Error(t, err)
	err = user.AddLimit("first pair", 0)
	assert.NoError(t, err)
	err = user.AddLimit("first pair", 1)
	assert.NoError(t, err)
}

func TestUser_GetLimit(t *testing.T) {
	setupUser()
	err := user.AddLimit("first pair", 1)
	assert.NoError(t, err)
	limit := user.GetLimit("first pair")
	assert.Equal(t, limit, 1.)
	err = user.AddLimit("first pair", 0.5)
	assert.NoError(t, err)
	limit = user.GetLimit("first pair")
	assert.Equal(t, limit, 0.5)
	limit = user.GetLimit("second pair")
	assert.Equal(t, limit, 0.)
	assert.NoError(t, err)
}
