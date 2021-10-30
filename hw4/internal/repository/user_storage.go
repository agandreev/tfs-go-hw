package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/agandreev/tfs-go-hw/hw4/internal/domain"
	"github.com/golang-jwt/jwt"
)

const (
	signingKey = "secret_signing_key"
	tokenTTL   = 10 * time.Minute
)

var (
	ErrUserExists    = errors.New("user with that username is already existed")
	ErrUserNotExists = errors.New("user with that username or password doesn't exist")
	ErrSigningMethod = errors.New("invalid signing method")
)

type AuthInMemory struct {
	rw sync.RWMutex
	Users map[string]*domain.User
}

type userClaims struct {
	jwt.StandardClaims
	UserID uint64 `json:"user_id"`
}

func (auth *AuthInMemory) CreateUser(user domain.User) (uint64, error) {
	auth.rw.RLock()
	if _, ok := auth.Users[user.Username]; ok {
		return 0, ErrUserExists
	}
	auth.rw.RUnlock()

	user.ID = uint64(len(auth.Users))
	auth.rw.Lock()
	auth.Users[user.Username] = &user
	auth.rw.Unlock()
	return user.ID, nil
}

func (auth *AuthInMemory) GenerateJWT(user domain.User) (string, error) {
	user, err := auth.GetUser(user.Username, user.Password)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &userClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.ID,
	})

	return token.SignedString([]byte(signingKey))
}

func (auth *AuthInMemory) ParseToken(accessToken string) (uint64, error) {
	token, err := jwt.ParseWithClaims(accessToken, &userClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrSigningMethod
			}
			return []byte(signingKey), nil
		})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*userClaims)
	if !ok {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserID, nil
}

func (auth *AuthInMemory) GetUser(username, password string) (domain.User, error) {
	auth.rw.RLock()
	if user, ok := auth.Users[username]; ok {
		if user.Password == password {
			return *user, nil
		}
	}
	auth.rw.RUnlock()
	return domain.User{}, ErrUserNotExists
}

func (auth *AuthInMemory) IsUserExist(id uint64) bool {
	auth.rw.RLock()
	for _, user := range auth.Users {
		if user.ID == id {
			return true
		}
	}
	auth.rw.Unlock()
	return false
}
