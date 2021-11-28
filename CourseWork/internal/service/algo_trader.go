package service

import (
	"fmt"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/repository/users"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/service/msgwriters"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// AlgoTrader describes trading service.
type AlgoTrader struct {
	Users             users.UserRepository
	Pairs             Pairs
	muPairs           *sync.Mutex
	API               StockMarketAPI
	MessageWriters    msgwriters.MessageWriters
	WG                *sync.WaitGroup
	reconnectionTimes int64
	signals           chan domain.StockMarketEvent
	errors            chan PairError
	stop              chan struct{}
	log               *logrus.Logger
}

// NewAlgoTrader returns pointer to AlgoTrader structure.
func NewAlgoTrader(users users.UserRepository, log *logrus.Logger,
	reconnections int64) *AlgoTrader {
	algoTrader := &AlgoTrader{
		Users:             users,
		Pairs:             make(Pairs),
		API:               NewKrakenAPI(),
		MessageWriters:    *msgwriters.NewMessageWriters(log),
		WG:                &sync.WaitGroup{},
		reconnectionTimes: reconnections,
		muPairs:           &sync.Mutex{},
		signals:           make(chan domain.StockMarketEvent),
		errors:            make(chan PairError),
		stop:              make(chan struct{}),
		log:               log,
	}
	return algoTrader
}

// Run runs infinite loop, which processes signal reading for order creating,
// and reads errors to reconnect sockets.
func (trader AlgoTrader) Run() {
	go func() {
		for {
			select {
			case event, ok := <-trader.signals:
				if !ok {
					break
				}
				trader.muPairs.Lock()
				for _, user := range trader.Pairs[event.Name][event.Interval].Users {
					message, err := trader.API.AddOrder(event, user)
					if err != nil {
						trader.log.Printf("ERROR: order is broken <%s>", err)
						trader.MessageWriters.WriteErrors(err.Error(), *user)
						continue
					}
					trader.MessageWriters.WriteMessages(message, *user)
				}
				trader.muPairs.Unlock()
			case pairErr := <-trader.errors:
				trader.WG.Done()
				pair := trader.Pairs[pairErr.Name][pairErr.Interval]
				err := trader.RunPair(pair)
				if err != nil {
					if len(pair.Users) == 0 {
						delete(trader.Pairs[pair.Name], pair.Interval)
						trader.log.Printf("DELETE: pair <%s> <%s> by error <%s>",
							pair.Name, pair.Interval, pairErr.Message)
						if len(trader.Pairs[pair.Name]) == 0 {
							delete(trader.Pairs, pair.Name)
							trader.log.Printf("DELETE: pair <%s> was deleted entirely by error <%s>",
								pair.Name, pairErr.Message)
						}
					}
					trader.MessageWriters.WriteErrorsToAll(err.Error(), pair.Users)
				}
			}
			trader.log.Println("STOP: all signals were processed gracefully")
			trader.stop <- struct{}{}
			break
		}
	}()
}

// ShutDown gracefully stops all running pairs.
func (trader AlgoTrader) ShutDown() {
	trader.muPairs.Lock()
	trader.Pairs.Shutdown(trader.WG)
	trader.muPairs.Unlock()
	trader.log.Println("STOP: All pairs were interrupted gracefully")
	close(trader.signals)
	close(trader.errors)
	<-trader.stop
	trader.MessageWriters.Shutdown()
}

// AddUser adds user to UserRepository.
func (trader AlgoTrader) AddUser(user domain.User) error {
	if err := trader.Users.AddUser(&user); err != nil {
		return err
	}
	trader.log.Printf("ADD: user: <%s> was added", user.Username)
	return nil
}

// AddPair adds pair and run it if it doesn't exist, else just sign user.
func (trader AlgoTrader) AddPair(username string, config domain.Config) error {
	if !config.Check() {
		return fmt.Errorf("incorrect config values")
	}
	user, err := trader.Users.GetUser(username)
	if err != nil {
		return err
	}
	trader.muPairs.Lock()
	if intervals, ok := trader.Pairs[config.PairName]; ok {
		if pair, ok := intervals[config.PairInterval]; ok {
			// pair name and pair interval are existing
			if err = pair.AddUser(user); err != nil {
				return err
			}
			trader.log.Printf("ADD: user <%s> was added to existed pair <%s> <%s>",
				username, pair.Name, pair.Interval)
		} else {
			// pair name is existing, but pair interval not
			if err = trader.AddAndRunPair(config.PairName, config.PairInterval,
				config.IndicatorName, user); err != nil {
				return err
			}
		}
	} else {
		trader.Pairs[config.PairName] = make(map[domain.CandleInterval]*Pair)
		// pair name is existing, but pair interval not
		if err = trader.AddAndRunPair(config.PairName, config.PairInterval,
			config.IndicatorName, user); err != nil {
			return err
		}
	}
	trader.muPairs.Unlock()
	return nil
}

// AddAndRunPair adds and runs pair.
func (trader AlgoTrader) AddAndRunPair(pairName string, pairInterval domain.CandleInterval,
	indicatorName string, user *domain.User) error {
	pair, err := NewPair(pairName, pairInterval, indicatorName, trader.log)
	if err != nil {
		return err
	}
	if err = trader.Pairs.AddPair(pair, user); err != nil {
		return err
	}
	trader.log.Printf("ADD: user <%s> was added and pair <%s> <%s> was created",
		user.Username, pair.Name, pair.Interval)
	if err = trader.RunPair(pair); err != nil {
		return err
	}
	trader.log.Printf("RUN: pair <%s> <%s> was run",
		pair.Name, pair.Interval)
	return nil
}

// RunPair runs pair and reconnect it if it's possible.
func (trader AlgoTrader) RunPair(pair *Pair) error {
	if err := pair.Run(trader.signals, trader.errors); err != nil {
		// reconnection
		ticker := time.NewTicker(5 * time.Second)
		var i int64
		for i = 0; i < trader.reconnectionTimes; i++ {
			<-ticker.C
			if err = pair.Run(trader.signals, trader.errors); err == nil {
				break
			}
		}
		if err != nil {
			return err
		}
	}
	trader.WG.Add(1)
	return nil
}

// DeletePair deletes pair if it is possible.
func (trader AlgoTrader) DeletePair(username string, config domain.Config) error {
	if !config.Check() {
		return fmt.Errorf("incorrect config values")
	}
	user, err := trader.Users.GetUser(username)
	if err != nil {
		return err
	}
	trader.muPairs.Lock()
	if intervals, ok := trader.Pairs[config.PairName]; ok {
		if pair, ok := intervals[config.PairInterval]; ok {
			if err = pair.DeleteUser(user); err != nil {
				return err
			}
			if len(pair.Users) == 0 {
				pair.Stop(trader.WG)
				delete(trader.Pairs[pair.Name], pair.Interval)
				trader.log.Printf("DELETE: pair <%s> <%s> was deleted by <%s>",
					pair.Name, pair.Interval, user.Username)
				if len(trader.Pairs[pair.Name]) == 0 {
					delete(trader.Pairs, pair.Name)
					trader.log.Printf("DELETE: pair <%s> was deleted entirely",
						pair.Name)
				}
			}
		}
	}
	trader.muPairs.Unlock()
	return nil
}

// AddMessageWriter adds message writer into slice.
func (trader *AlgoTrader) AddMessageWriter(messageWriter msgwriters.MessageWriter) {
	trader.MessageWriters.AddWriter(messageWriter)
}
