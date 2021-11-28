package msgwriters

import (
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/sirupsen/logrus"
)

type MessageWriter interface {
	WriteMessage(message domain.Message, user domain.User) error
	WriteError(message string, user domain.User) error
	Shutdown()
}

type MessageWriters struct {
	Writers []MessageWriter
	log *logrus.Logger
}

func NewMessageWriters(log *logrus.Logger) *MessageWriters {
	return &MessageWriters{
		Writers: make([]MessageWriter, 0),
		log:     log,
	}
}

func (writers *MessageWriters) AddWriter(writer MessageWriter) {
	writers.Writers = append(writers.Writers, writer)
}

func (writers MessageWriters) WriteMessages(message domain.Message, user domain.User) {
	for _, writer := range writers.Writers {
		if err := writer.WriteMessage(message, user); err != nil {
			writers.log.Printf("WRITE: <%s>", err)
			continue
		}
	}
}

func (writers MessageWriters) WriteErrors(message string, user domain.User) {
	for _, writer := range writers.Writers {
		if err := writer.WriteError(message, user); err != nil {
			writers.log.Printf("WRITE: <%s>", err)
			continue
		}
	}
}

func (writers MessageWriters) WriteErrorsToAll(message string, users []*domain.User) {
	for _, user := range users {
		for _, writer := range writers.Writers {
			if err := writer.WriteError(message, *user); err != nil {
				writers.log.Printf("WRITE: <%s>", err)
				continue
			}
		}
	}
}
