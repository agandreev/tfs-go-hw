package service

import (
	"sync"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/sirupsen/logrus"
)

type MessageWriters struct {
	Writers   []MessageWriter
	muWriters *sync.Mutex
	log       *logrus.Logger
}

func NewMessageWriters(log *logrus.Logger) *MessageWriters {
	return &MessageWriters{
		Writers:   make([]MessageWriter, 0),
		muWriters: &sync.Mutex{},
		log:       log,
	}
}

func (writers *MessageWriters) AddWriter(writer MessageWriter) {
	writers.muWriters.Lock()
	writers.Writers = append(writers.Writers, writer)
	writers.muWriters.Unlock()
}

func (writers MessageWriters) WriteMessages(message domain.OrderInfo, user domain.User) {
	writers.muWriters.Lock()
	for _, writer := range writers.Writers {
		if err := writer.WriteMessage(message, user); err != nil {
			writers.log.Printf("WRITE: <%s>", err)
			continue
		}
	}
	writers.muWriters.Unlock()
}

func (writers MessageWriters) WriteErrors(message string, user domain.User) {
	writers.muWriters.Lock()
	for _, writer := range writers.Writers {
		if err := writer.WriteError(message, user); err != nil {
			writers.log.Printf("WRITE: <%s>", err)
			continue
		}
	}
	writers.muWriters.Unlock()
}

func (writers MessageWriters) WriteErrorsToAll(message string, users []*domain.User) {
	writers.muWriters.Lock()
	for _, user := range users {
		for _, writer := range writers.Writers {
			if err := writer.WriteError(message, *user); err != nil {
				writers.log.Printf("WRITE: <%s>", err)
				continue
			}
		}
	}
	writers.muWriters.Unlock()
}

func (writers MessageWriters) Shutdown() {
	writers.muWriters.Lock()
	for _, writer := range writers.Writers {
		writer.Shutdown()
	}
	writers.muWriters.Unlock()
}
