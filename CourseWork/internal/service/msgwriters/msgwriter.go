package msgwriters

import (
	"fmt"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
)

type MessageWriter interface {
	WriteMessage(message domain.Message, user domain.User) error
}

type ErrorWriter interface {
	WriteError(message string, user domain.User) error
}

type ConsoleWriter struct {

}

func (console ConsoleWriter) WriteMessage(message domain.Message, user domain.User) error {
	fmt.Println(message)
	return nil
}

func (console ConsoleWriter) WriteError(message string, user domain.User) error {
	fmt.Println(message)
	return nil
}

type MessageWriters []MessageWriter

func (writers MessageWriters) WriteMessage(message domain.Message, user domain.User) {
	for _, writer := range writers {
		err := writer.WriteMessage(message, user)
		if err != nil {
			continue
		}
	}
}

func (writers MessageWriters) WriteError(message string, user domain.User) {
	for _, writer := range writers {
		if errorWriter, ok := writer.(ErrorWriter); ok {
			if err := errorWriter.WriteError(message, user); err != nil {
				continue
			}
		}
	}
}

func (writers MessageWriters) WriteErrorToAll(message string, users []*domain.User) {
	for _, user := range users {
		for _, writer := range writers {
			if errorWriter, ok := writer.(ErrorWriter); ok {
				if err := errorWriter.WriteError(message, *user); err != nil {
					continue
				}
			}
		}
	}
}
