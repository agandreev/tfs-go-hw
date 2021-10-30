package repository

import (
	"errors"
	"sync"

	"github.com/agandreev/tfs-go-hw/hw4/internal/domain"
)

var (
	ErrEmptyStorage = errors.New("storage is empty")
	ErrIncorrectID  = errors.New("id is too large")
)

type MessageStorage struct {
	generalRW       sync.RWMutex
	GeneralMessages []*domain.GeneralMessage
	directRW        sync.RWMutex
	DirectMessages  map[uint64][]*domain.DirectMessage
}

func (storage *MessageStorage) AddGeneral(message domain.GeneralMessage) {
	storage.generalRW.Lock()
	message.ID = uint64(len(storage.GeneralMessages))
	storage.GeneralMessages = append(storage.GeneralMessages, &message)
	storage.generalRW.Unlock()
}

func (storage *MessageStorage) GetGeneral() (domain.GeneralMessage, error) {
	storage.generalRW.RLock()
	defer storage.generalRW.RUnlock()
	if len(storage.GeneralMessages) == 0 {
		return domain.GeneralMessage{}, ErrEmptyStorage
	}
	return *(storage.GeneralMessages)[len(storage.GeneralMessages)-1], nil
}

func (storage *MessageStorage) ListGeneral(offset, messageID uint64) (
	[]*domain.GeneralMessage, error) {
	storage.directRW.RLock()
	defer storage.directRW.Unlock()
	if messageID >= uint64(len(storage.GeneralMessages)) {
		return nil, ErrIncorrectID
	}
	if offset > messageID {
		return (storage.GeneralMessages)[:messageID+1], nil
	}
	return (storage.GeneralMessages)[messageID-offset : messageID+1], nil
}

func (storage *MessageStorage) AddDirect(message domain.DirectMessage) {
	storage.directRW.Lock()
	if storage.DirectMessages[message.ReceiverID] == nil {
		messages := make([]*domain.DirectMessage, 0)
		storage.DirectMessages[message.ReceiverID] = messages
	}
	message.ID = uint64(len(storage.DirectMessages[message.ReceiverID]))
	storage.DirectMessages[message.ReceiverID] =
		append(storage.DirectMessages[message.ReceiverID], &message)
	storage.directRW.Unlock()
}
func (storage *MessageStorage) ListDirect(userID, offset, messageID uint64) (
	[]*domain.DirectMessage, error) {
	storage.directRW.RLock()
	defer storage.directRW.Unlock()
	if messageID >= uint64(len(storage.DirectMessages[userID])) {
		return nil, ErrIncorrectID
	}
	if offset > messageID {
		return (storage.DirectMessages[userID])[messageID:], nil
	}
	return (storage.DirectMessages[userID])[messageID-offset : messageID+1], nil
}
func (storage *MessageStorage) GetDirect(userID uint64) (domain.DirectMessage, error) {
	storage.directRW.RLock()
	defer storage.directRW.Unlock()
	if len(storage.DirectMessages[userID]) == 0 {
		return domain.DirectMessage{}, ErrEmptyStorage
	}
	return *(storage.DirectMessages[userID])[len(storage.DirectMessages[userID])-1], nil
}
