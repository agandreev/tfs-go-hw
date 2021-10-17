package main

import (
	"github.com/agandreev/tfs-go-hw/hw4/internal/controller"
	"github.com/agandreev/tfs-go-hw/hw4/internal/domain"
	"github.com/agandreev/tfs-go-hw/hw4/internal/repository"
	"github.com/agandreev/tfs-go-hw/hw4/internal/service"
)

func main() {
	generalMessages := make([]*domain.GeneralMessage, 0)
	server := controller.Server{
		GeneralChat: service.GeneralChat{
			Users: repository.AuthInMemory{
				Users: make(map[string]*domain.User)},
			Messages: &repository.MessageStorage{
				GeneralMessages: &generalMessages,
				DirectMessages:  make(map[uint64]*[]*domain.DirectMessage),
			}}}
	server.Run()
}
