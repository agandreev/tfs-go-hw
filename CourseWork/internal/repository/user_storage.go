package repository

import (
	"errors"
	"fmt"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/golang-jwt/jwt"
	"sync"
	"time"
)

var (
	ErrExistedUser         = errors.New("the user is already existed")
	ErrNonExistentUser     = errors.New("the user does not exist")
	ErrIncorrectUserValues = errors.New("the user's id, public key or private key is empty")
	ErrNothingToChange     = errors.New("nothing to change or already changed")
)

// UserStorage implements UserRepository.
type UserStorage struct {
	muUsers  *sync.RWMutex
	users    map[string]*domain.User
	signKey  string
	ttlHours int64
}

// NewUserStorage returns pointer to UserStorage.
func NewUserStorage(signKey string, ttlHours int64) (*UserStorage, error) {
	if len(signKey) == 0 {
		return nil, fmt.Errorf("sign key is too short")
	}
	if ttlHours <= 0 {
		return nil, fmt.Errorf("ttl hourse should be more than zero")
	}
	storage := &UserStorage{users: make(map[string]*domain.User),
		muUsers:  &sync.RWMutex{},
		signKey:  signKey,
		ttlHours: ttlHours}
	return storage, nil
}

// userClaims is a part of JWT Token generation.
type userClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
}

// AddUser adds user to storage and checks it for errors.
func (u UserStorage) AddUser(user *domain.User) error {
	if len(user.Username) == 0 || len(user.PublicKey) == 0 || len(user.PrivateKey) == 0 {
		return ErrIncorrectUserValues
	}
	u.muUsers.Lock()
	defer u.muUsers.Unlock()
	if _, ok := u.users[user.Username]; ok {
		return ErrExistedUser
	}
	u.users[user.Username] = user
	return nil
}

// DeleteUser delete user from storage.
func (u UserStorage) DeleteUser(username string) error {
	u.muUsers.Lock()
	defer u.muUsers.Unlock()
	if _, ok := u.users[username]; !ok {
		return ErrNonExistentUser
	}
	delete(u.users, username)
	return nil
}

// GetUser returns User from storage.
func (u UserStorage) GetUser(username string) (*domain.User, error) {
	u.muUsers.RLock()
	defer u.muUsers.RUnlock()
	if user, ok := u.users[username]; !ok {
		return nil, ErrNonExistentUser
	} else {
		return user, nil
	}
}

// SetKeys edits user's keys.
func (u UserStorage) SetKeys(username, public, private string) error {
	if len(username) == 0 || len(public) == 0 || len(private) == 0 {
		return ErrIncorrectUserValues
	}
	user, err := u.GetUser(username)
	if err != nil {
		return err
	}
	if user.PublicKey == public && user.PrivateKey == private {
		return ErrNothingToChange
	}
	user.PublicKey = public
	user.PrivateKey = private
	return nil
}

// GenerateJWT generates JWT token.
func (u UserStorage) GenerateJWT(user domain.User) (string, error) {
	storageUser, err := u.GetUser(user.Username)
	if err != nil {
		return "", err
	}
	if storageUser.PublicKey != user.PublicKey ||
		storageUser.PrivateKey != user.PrivateKey {
		return "", fmt.Errorf("incorrect public or private key")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &userClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(u.ttlHours) * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.Username,
	})

	return token.SignedString([]byte(u.signKey))
}

//ParseToken parses JWT token.
func (u UserStorage) ParseToken(accessToken string) (string, error) {
	token, err := jwt.ParseWithClaims(accessToken, &userClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return []byte(u.signKey), nil
		})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*userClaims)
	if !ok {
		return "", fmt.Errorf("token claims are not of type *tokenClaims")
	}

	return claims.Username, nil
}
