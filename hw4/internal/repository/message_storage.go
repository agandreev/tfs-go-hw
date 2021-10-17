package repository

import (
	"errors"

	"github.com/agandreev/tfs-go-hw/hw4/internal/domain"
)

var (
	ErrEmptyStorage = errors.New("storage is empty")
	ErrIncorrectID  = errors.New("id is too large")
)

type MessageStorage struct {
	GeneralMessages *[]*domain.GeneralMessage
	DirectMessages  map[uint64]*[]*domain.DirectMessage
}

func (storage *MessageStorage) AddGeneral(message domain.GeneralMessage) {
	message.ID = uint64(len(*storage.GeneralMessages))
	*storage.GeneralMessages = append(*storage.GeneralMessages, &message)
}

func (storage MessageStorage) GetGeneral() (domain.GeneralMessage, error) {
	if len(*storage.GeneralMessages) == 0 {
		return domain.GeneralMessage{}, ErrEmptyStorage
	}
	return *(*storage.GeneralMessages)[len(*storage.GeneralMessages)-1], nil
}

func (storage MessageStorage) ListGeneral(offset, messageID uint64) (
	[]*domain.GeneralMessage, error) {
	if messageID >= uint64(len(*storage.GeneralMessages)) {
		return nil, ErrIncorrectID
	}
	if offset > messageID {
		return (*storage.GeneralMessages)[:messageID+1], nil
	}
	return (*storage.GeneralMessages)[messageID-offset : messageID+1], nil
}

func (storage *MessageStorage) AddDirect(message domain.DirectMessage) {
	if storage.DirectMessages[message.ReceiverID] == nil {
		messages := make([]*domain.DirectMessage, 0)
		storage.DirectMessages[message.ReceiverID] = &messages
	}
	message.ID = uint64(len(*storage.DirectMessages[message.ReceiverID]))
	*storage.DirectMessages[message.ReceiverID] =
		append(*storage.DirectMessages[message.ReceiverID], &message)
}
func (storage MessageStorage) ListDirect(userID, offset, messageID uint64) (
	[]*domain.DirectMessage, error) {
	if messageID >= uint64(len(*storage.DirectMessages[userID])) {
		return nil, ErrIncorrectID
	}
	if offset > messageID {
		return (*storage.DirectMessages[userID])[messageID:], nil
	}
	return (*storage.DirectMessages[userID])[messageID-offset : messageID+1], nil
}
func (storage MessageStorage) GetDirect(userID uint64) (domain.DirectMessage, error) {
	if len(*storage.DirectMessages[userID]) == 0 {
		return domain.DirectMessage{}, ErrEmptyStorage
	}
	return *(*storage.DirectMessages[userID])[len(*storage.DirectMessages[userID])-1], nil
}
