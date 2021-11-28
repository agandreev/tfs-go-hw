package service

import (
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPair_AddUser(t *testing.T) {
	pair, err := NewPair("name", "", DonchianName, nil)
	user := domain.NewUser("name")
	err = pair.AddUser(user)
	assert.NoError(t, err)
	assert.Equal(t, len(pair.Users), 1)

	err = pair.AddUser(user)
	assert.Error(t, err)
	assert.Equal(t, len(pair.Users), 1)
}

func TestPair_DeleteUser(t *testing.T) {
	pair, err := NewPair("name", "", DonchianName, nil)
	user := domain.NewUser("name")
	_ = pair.AddUser(user)
	err = pair.DeleteUser(user)
	assert.NoError(t, err)
	assert.Equal(t, len(pair.Users), 0)

	err = pair.DeleteUser(user)
	assert.Error(t, err)
}

func TestPair_IsUserLogged(t *testing.T) {
	pair, _ := NewPair("name", "", DonchianName, nil)
	user := domain.NewUser("name")
	assert.False(t, pair.IsUserLogged(user))
	_ = pair.AddUser(user)
	assert.True(t, pair.IsUserLogged(user))
	_ = pair.DeleteUser(user)
	assert.False(t, pair.IsUserLogged(user))
}
