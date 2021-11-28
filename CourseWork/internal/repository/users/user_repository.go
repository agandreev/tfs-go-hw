package users

import (
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
)

// UserRepository describes UserStorage methods.
type UserRepository interface {
	AddUser(user *domain.User) error
	DeleteUser(string) error
	GetUser(string) (*domain.User, error)
	SetKeys(string, string, string) error
	GenerateJWT(domain.User) (string, error)
	ParseToken(string) (string, error)
}
