package service

//go:generate mockgen -source=pair.go -destination=pair_mock.go

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/sirupsen/logrus"
)

const (
	DonchianName = "Donchian"
)

var (
	ErrUserIsLogged    = errors.New("current user is already logged")
	ErrUserIsNotLogged = errors.New("current user is not logged")
)

type StockMarketSocket interface {
	Connect(context.Context) error
	SubscribeCandle(Pair, chan domain.Candle, chan PairError) error
}

// Indicator implements strategy of stock market service.
type Indicator interface {
	Add(candle domain.Candle) (domain.Signal, error)
}

// Pair describes stock market pair entity.
type Pair struct {
	Name      string
	Users     []*domain.User
	Interval  domain.CandleInterval
	Indicator Indicator
	stop      chan struct{}
	socket    StockMarketSocket
	ctx       context.Context
	cancel    context.CancelFunc
	log       *logrus.Logger
}

// NewPair returns pointer to Pair with custom parameters.
func NewPair(name string, interval domain.CandleInterval, indicatorName string,
	log *logrus.Logger) (*Pair, error) {
	pair := &Pair{
		Name:      name,
		Users:     make([]*domain.User, 0),
		Interval:  interval,
		Indicator: domain.NewDonchian(),
		stop:      make(chan struct{}),
		socket:    &KrakenSocket{log: log},
		ctx:       context.Background(),
		log:       log,
	}
	switch indicatorName {
	case DonchianName:
		pair.Indicator = domain.NewDonchian()
	default:
		return nil, fmt.Errorf("such indicator doesn't exist")
	}
	return pair, nil
}

// AddUser subscribes user to current Pair.
func (pair *Pair) AddUser(user *domain.User) error {
	if pair.IsUserLogged(user) {
		return ErrUserIsLogged
	}
	pair.Users = append(pair.Users, user)
	return nil
}

// DeleteUser unsubscribes user from current Pair.
func (pair *Pair) DeleteUser(user *domain.User) error {
	if !pair.IsUserLogged(user) {
		return ErrUserIsNotLogged
	}
	var err error
	if pair.Users, err = remove(pair.Users, user); err != nil {
		return err
	}
	pair.log.Printf("REMOVE: user <%s> was removed from pair <%s> <%s>",
		user.Username, pair.Name, pair.Interval)
	return nil
}

// Stop gracefully shutdowns current Pair.
func (pair Pair) Stop(wg *sync.WaitGroup) {
	pair.cancel()
	<-pair.stop
	wg.Done()
}

// IsUserLogged cheks if User is subscribed.
func (pair Pair) IsUserLogged(user *domain.User) bool {
	for _, loggedUser := range pair.Users {
		if user.Username == loggedUser.Username {
			return true
		}
	}
	return false
}

// Run represents second step of pipeline where
// there is a process of sending candles to Indicator.
func (pair *Pair) Run(events chan domain.StockMarketEvent, errors chan PairError) error {
	ctx, cancel := context.WithCancel(pair.ctx)
	pair.cancel = cancel
	candles := make(chan domain.Candle)
	if err := pair.socket.Connect(ctx); err != nil {
		return fmt.Errorf("can't connect <%w>", err)
	}
	err := pair.socket.SubscribeCandle(*pair, candles, errors)
	if err != nil {
		return fmt.Errorf("can't sucscribe candle <%w>", err)
	}
	go func() {
		for candle := range candles {
			signal, err := pair.Indicator.Add(candle)
			if err != nil {
				// timestamp checking
				if err == domain.ErrSameTimestamp {
					continue
				}
				// other error sending
				errors <- PairError{
					Name:     pair.Name,
					Interval: pair.Interval,
					Message:  err.Error(),
				}
				return
			}
			pair.log.Printf("%s Signal: %s", candle, signal)
			if signal == domain.Buy || signal == domain.Sell {
				events <- domain.StockMarketEvent{
					Signal:   signal,
					Name:     pair.Name,
					Interval: pair.Interval,
					Volume:   candle.Volume,
					Close:    candle.Close,
				}
			}
		}
		pair.log.Printf("STOP: <%s> <%s> was interrupted gracefully", pair.Name, pair.Interval)
		pair.stop <- struct{}{}
	}()
	return nil
}

// remove deletes user from user's slice.
func remove(users []*domain.User, deletingUser *domain.User) ([]*domain.User, error) {
	for i, user := range users {
		if user == deletingUser {
			return append(users[:i], users[i+1:]...), nil
		}
	}
	return nil, fmt.Errorf("user don't exist")
}

// Pairs represents map of Pairs by Pair's name and Pair's interval.
type Pairs map[string]map[domain.CandleInterval]*Pair

// IsExist checks if pair with such pairName and interval exist or not.
func (pairs Pairs) IsExist(pairName string, interval domain.CandleInterval) bool {
	if intervals, ok := pairs[pairName]; ok {
		if _, ok = intervals[interval]; ok {
			return true
		}
	}
	return false
}

// Shutdown gracefully shutdowns all running pairs.
func (pairs Pairs) Shutdown(wg *sync.WaitGroup) {
	for _, interval := range pairs {
		for _, pair := range interval {
			go pair.Stop(wg)
		}
	}
	wg.Wait()
}

// AddPair adds Pair to nested pair's map.
func (pairs Pairs) AddPair(pair *Pair, user *domain.User) error {
	pairs[pair.Name][pair.Interval] = pair
	if err := pair.AddUser(user); err != nil {
		return err
	}
	return nil
}

// PairError represents error connected to definite pair.
type PairError struct {
	Name     string
	Interval domain.CandleInterval
	Message  string
}

func (pairError PairError) Error() string {
	return pairError.Message
}
