package repository

import "github.com/agandreev/tfs-go-hw/hw4/internal/domain"

type MessageRepository interface {
	AddGeneral(message domain.GeneralMessage)
	ListGeneral(limit, messageID uint64) ([]*domain.GeneralMessage, error)
	GetGeneral() (domain.GeneralMessage, error)

	AddDirect(message domain.DirectMessage)
	ListDirect(userID, limit, messageID uint64) ([]*domain.DirectMessage, error)
	GetDirect(userID uint64) (domain.DirectMessage, error)
}

type UserRepository interface {
	GetUser(name, password string) (domain.User, error)
	ParseToken(accessToken string) (uint64, error)
	GenerateJWT(user domain.User) (string, error)
	CreateUser(user domain.User) (uint64, error)
	IsUserExist(id uint64) bool
}
