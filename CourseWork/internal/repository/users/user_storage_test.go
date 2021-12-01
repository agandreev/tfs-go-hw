package users

import (
	"testing"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/stretchr/testify/assert"
)

var (
	user = &domain.User{
		Username:   "name",
		PublicKey:  "0",
		PrivateKey: "0",
		TelegramID: 0,
	}
	emptyString = ""
)

func TestUserStorage_NewUserStorage(t *testing.T) {
	storage, err := NewUserStorage("", 1)
	assert.Nil(t, storage)
	assert.Error(t, err)

	storage, err = NewUserStorage("x", 0)
	assert.Nil(t, storage)
	assert.Error(t, err)
}

func TestUserStorage_AddUser(t *testing.T) {
	storage, err := NewUserStorage("key", 1)
	assert.NoError(t, err)
	err = storage.AddUser(&domain.User{})
	assert.Error(t, err)
	err = storage.AddUser(&domain.User{Username: "name"})
	assert.ErrorIs(t, err, ErrIncorrectUserValues)
	err = storage.AddUser(&domain.User{PublicKey: "name"})
	assert.ErrorIs(t, err, ErrIncorrectUserValues)
	err = storage.AddUser(&domain.User{PrivateKey: "name"})
	assert.ErrorIs(t, err, ErrIncorrectUserValues)
	err = storage.AddUser(user)
	assert.NoError(t, err)
	assert.Equal(t, storage.users[user.Username], user)
	err = storage.AddUser(user)
	assert.Error(t, err)
}

func TestUserStorage_DeleteUser(t *testing.T) {
	storage, err := NewUserStorage("key", 1)
	assert.NoError(t, err)
	_ = storage.AddUser(user)
	err = storage.DeleteUser(user.Username)
	assert.NoError(t, err)
	assert.NotEqual(t, storage.users[user.Username], user)
	err = storage.DeleteUser(user.Username)
	assert.Error(t, err)
}

func TestUserStorage_GetUser(t *testing.T) {
	storage, err := NewUserStorage("key", 1)
	assert.NoError(t, err)
	_ = storage.AddUser(user)
	storageUser, err := storage.GetUser(user.Username)
	assert.NoError(t, err)
	assert.Equal(t, user, storageUser)

	storageUser, err = storage.GetUser("not a user")
	assert.Error(t, err)
	assert.Nil(t, storageUser)
}

func TestUserStorage_SetKeys(t *testing.T) {
	storage, err := NewUserStorage("key", 1)
	assert.NoError(t, err)
	_ = storage.AddUser(user)
	err = storage.SetKeys(emptyString, emptyString, emptyString)
	assert.ErrorIs(t, err, ErrIncorrectUserValues)
	err = storage.SetKeys(user.Username, user.PublicKey, user.PrivateKey)
	assert.Error(t, err)
	err = storage.SetKeys(user.Username, user.PublicKey, user.PrivateKey)
	assert.ErrorIs(t, err, ErrNothingToChange)
	err = storage.SetKeys(user.Username, user.PublicKey+"0", user.PrivateKey+"0")
	assert.NoError(t, err)
	err = storage.SetKeys(user.Username+"x", user.PublicKey, user.PrivateKey)
	assert.ErrorIs(t, err, ErrNonExistentUser)
}

func TestUserStorage_GenerateJWT(t *testing.T) {
	storage, err := NewUserStorage("key", 1)
	assert.NoError(t, err)
	_ = storage.AddUser(user)
	jwt, err := storage.GenerateJWT(domain.User{Username: "x"})
	assert.Error(t, err)
	assert.Empty(t, jwt)

	jwt, err = storage.GenerateJWT(domain.User{Username: user.Username})
	assert.Error(t, err)
	assert.Empty(t, jwt)

	jwt, err = storage.GenerateJWT(*user)
	assert.NoError(t, err)
	username, err := storage.ParseToken(jwt)
	assert.NoError(t, err)
	assert.Equal(t, username, user.Username)
}
