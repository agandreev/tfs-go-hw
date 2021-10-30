package service

import (
	"github.com/agandreev/tfs-go-hw/hw4/internal/domain"
	"github.com/agandreev/tfs-go-hw/hw4/internal/repository"

	"fmt"
)

type GeneralChat struct {
	Users    repository.UserRepository
	Messages repository.MessageRepository
}

func (chat GeneralChat) SendDirectMessage(message domain.DirectMessage) error {
	if !chat.Users.IsUserExist(message.ReceiverID) {
		return fmt.Errorf("receiver doesn't exist")
	}
	chat.Messages.AddDirect(message)
	return nil
}
