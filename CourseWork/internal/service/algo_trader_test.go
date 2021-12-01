package service

import (
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/repository/users"
	mock_service "github.com/agandreev/tfs-go-hw/CourseWork/internal/service/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
)

var trader = setupTrader()

func setupTrader() *AlgoTrader {
	log := logrus.New()
	trader := &AlgoTrader{
		Users:             nil,
		Pairs:             make(Pairs),
		Orders:            nil,
		API:               NewKrakenAPI(),
		MessageWriters:    *NewMessageWriters(log),
		wg:                &sync.WaitGroup{},
		reconnectionTimes: 1,
		muPairs:           &sync.Mutex{},
		signals:           make(chan domain.StockMarketEvent),
		errors:            make(chan PairError),
		stop:              make(chan struct{}),
		log:               log,
	}
	trader.log.SetOutput(io.Discard)
	return trader
}

func TestAlgoTrader_AddUser(t *testing.T) {
	type mockBehavior func(r *mock_service.MockUserRepository, user *domain.User)

	tests := []struct {
		name          string
		inputUser     domain.User
		mockBehavior  mockBehavior
		expectedValue error
	}{
		{
			name: "Ok",
			inputUser: domain.User{
				Username:   "username",
				PublicKey:  "x",
				PrivateKey: "x",
			},
			mockBehavior: func(r *mock_service.MockUserRepository, user *domain.User) {
				r.EXPECT().AddUser(user).Return(nil)
			},
			expectedValue: nil,
		},
		{
			name: "Ok with tg",
			inputUser: domain.User{
				Username:   "username",
				PublicKey:  "x",
				PrivateKey: "x",
				TelegramID: 1,
			},
			mockBehavior: func(r *mock_service.MockUserRepository, user *domain.User) {
				r.EXPECT().AddUser(user).Return(nil)
			},
			expectedValue: nil,
		},
		{
			name: "less len",
			inputUser: domain.User{
				Username:   "",
				PublicKey:  "x",
				PrivateKey: "x",
				TelegramID: 1,
			},
			mockBehavior: func(r *mock_service.MockUserRepository, user *domain.User) {
				r.EXPECT().AddUser(user).Return(users.ErrIncorrectUserValues)
			},
			expectedValue: users.ErrIncorrectUserValues,
		},
		{
			name: "less len",
			inputUser: domain.User{
				Username:   "x",
				PublicKey:  "",
				PrivateKey: "x",
				TelegramID: 1,
			},
			mockBehavior: func(r *mock_service.MockUserRepository, user *domain.User) {
				r.EXPECT().AddUser(user).Return(users.ErrIncorrectUserValues)
			},
			expectedValue: users.ErrIncorrectUserValues,
		},
		{
			name: "less len",
			inputUser: domain.User{
				Username:   "x",
				PublicKey:  "x",
				PrivateKey: "",
				TelegramID: 1,
			},
			mockBehavior: func(r *mock_service.MockUserRepository, user *domain.User) {
				r.EXPECT().AddUser(user).Return(users.ErrIncorrectUserValues)
			},
			expectedValue: users.ErrIncorrectUserValues,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_service.NewMockUserRepository(c)
			test.mockBehavior(repo, &test.inputUser)

			trader.Users = repo

			// Assert
			if test.expectedValue == nil {
				assert.Equal(t, trader.AddUser(test.inputUser), test.expectedValue)
			} else {
				assert.Error(t, trader.AddUser(test.inputUser))
			}
		})
	}
}
