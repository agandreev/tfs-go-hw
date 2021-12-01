package service

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	pair  = makePair()
	pairs = make(Pairs)
)

func TestMain(m *testing.M) {
	pair.log.SetOutput(io.Discard)
	code := m.Run()
	os.Exit(code)
}

func setup() {
	pair = makePair()
}

func setupPairs() {
	pairs = make(Pairs)
}

func makePair() *Pair {
	return &Pair{
		Name:      "name",
		Users:     make([]*domain.User, 0),
		Interval:  domain.Candle1m,
		Indicator: &MockIndicator{},
		stop:      make(chan struct{}),
		socket:    &MockStockMarketSocket{},
		ctx:       context.Background(),
		log:       logrus.New(),
	}
}

func TestNewPair(t *testing.T) {
	nilPair, err := NewPair("name", "", DonchianName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, nilPair)

	nilPair, err = NewPair("name", "", "", nil)
	assert.Error(t, err)
	assert.Nil(t, nilPair)
}

func TestPair_AddUser(t *testing.T) {
	setup()
	user := domain.NewUser("name")
	err := pair.AddUser(user)
	assert.NoError(t, err)
	assert.Equal(t, len(pair.Users), 1)

	err = pair.AddUser(user)
	assert.Error(t, err)
	assert.Equal(t, len(pair.Users), 1)
}

func TestPair_DeleteUser(t *testing.T) {
	setup()
	user := domain.NewUser("name")
	_ = pair.AddUser(user)
	err := pair.DeleteUser(user)
	assert.NoError(t, err)
	assert.Equal(t, len(pair.Users), 0)

	err = pair.DeleteUser(user)
	assert.Error(t, err)
}

func TestPair_IsUserLogged(t *testing.T) {
	setup()
	user := domain.NewUser("name")
	assert.False(t, pair.IsUserLogged(user))
	_ = pair.AddUser(user)
	assert.True(t, pair.IsUserLogged(user))
	_ = pair.DeleteUser(user)
	assert.False(t, pair.IsUserLogged(user))
}

func TestPairs_AddPair(t *testing.T) {
	user := domain.NewUser("name")
	pairs[pair.Name] = make(map[domain.CandleInterval]*Pair)
	err := pairs.AddPair(pair, user)
	assert.NoError(t, err)
	assert.Equal(t, len(pairs[pair.Name]), 1)
	err = pairs.AddPair(pair, user)
	assert.Error(t, err)
	assert.Equal(t, len(pairs[pair.Name]), 1)
}

func TestPairs_IsExist(t *testing.T) {
	setupPairs()
	assert.False(t, pairs.IsExist(pair.Name, pair.Interval))
	user := domain.NewUser("name")
	pairs[pair.Name] = make(map[domain.CandleInterval]*Pair)
	_ = pairs.AddPair(pair, user)
	assert.True(t, pairs.IsExist(pair.Name, pair.Interval))
}
