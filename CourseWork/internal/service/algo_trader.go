package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -source=algo_trader.go -destination=mocks/mock.go

// UserRepository describes UserStorage methods.
type UserRepository interface {
	AddUser(user *domain.User) error
	DeleteUser(string) error
	GetUser(string) (*domain.User, error)
	SetKeys(string, string, string) error
	GenerateJWT(domain.User) (string, error)
	ParseToken(string) (string, error)
}

type OrderRepository interface {
	AddOrder(domain.OrderInfo) error
	GetOrders(int64) ([]domain.OrderInfo, error)
	Connect() error
	Shutdown()
}

type MessageWriter interface {
	WriteMessage(message domain.OrderInfo, user domain.User) error
	WriteError(message string, user domain.User) error
	Shutdown()
}

type StockMarketAPI interface {
	AddOrder(domain.StockMarketEvent, *domain.User) (domain.OrderInfo, error)
}

// AlgoTrader describes trading service.
type AlgoTrader struct {
	Users             UserRepository
	Pairs             Pairs
	Orders            OrderRepository
	muPairs           *sync.Mutex
	API               StockMarketAPI
	MessageWriters    MessageWriters
	WG                *sync.WaitGroup
	reconnectionTimes int64
	signals           chan domain.StockMarketEvent
	errors            chan PairError
	stop              chan struct{}
	log               *logrus.Logger
}

// NewAlgoTrader returns pointer to AlgoTrader structure.
func NewAlgoTrader(users UserRepository, orders OrderRepository,
	log *logrus.Logger, reconnections int64) *AlgoTrader {
	algoTrader := &AlgoTrader{
		Users:             users,
		Pairs:             make(Pairs),
		Orders:            orders,
		API:               NewKrakenAPI(),
		MessageWriters:    *NewMessageWriters(log),
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
func (trader AlgoTrader) Run() error {
	err := trader.Orders.Connect()
	if err != nil {
		return fmt.Errorf("error while try to run <%w>", err)
	}
	go func() {
		isOut := false
		for {
			select {
			case event, ok := <-trader.signals:
				if !ok {
					isOut = true
					break
				}
				trader.stockEventHandler(event)
			case pairErr := <-trader.errors:
				trader.stockErrorHandler(pairErr)
			}
			if isOut {
				break
			}
		}
		trader.log.Println("STOP: all signals were processed gracefully")
		trader.stop <- struct{}{}
	}()
	return nil
}

// stockEventHandler process Indicator signals and transfer information.
func (trader AlgoTrader) stockEventHandler(event domain.StockMarketEvent) {
	trader.muPairs.Lock()
	defer trader.muPairs.Unlock()
	for _, user := range trader.Pairs[event.Name][event.Interval].Users {
		orderInfo, err := trader.API.AddOrder(event, user)
		if err != nil {
			trader.log.Printf("ERROR: order is broken <%s>", err)
			eventError := domain.StockMarketEventError{Event: event,
				ErrorMessage: err.Error()}
			trader.MessageWriters.WriteErrors(eventError.String(), *user)
			continue
		}
		trader.MessageWriters.WriteMessages(orderInfo, *user)
		if err = trader.Orders.AddOrder(orderInfo); err != nil {
			trader.log.Printf("DB: <%s>", err)
			continue
		}
		trader.log.Println("DB: order is created successfully")
	}
}

// stockErrorHandler tries to recover dead pair and delete it.
func (trader AlgoTrader) stockErrorHandler(pairErr PairError) {
	trader.muPairs.Lock()
	defer trader.muPairs.Unlock()
	trader.WG.Done()
	pair := trader.Pairs[pairErr.Name][pairErr.Interval]
	// try to recover pair or delete it
	if err := trader.RunPair(pair); err != nil {
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

// ShutDown gracefully stops all running pairs.
func (trader AlgoTrader) ShutDown() {
	trader.muPairs.Lock()
	trader.Pairs.Shutdown(trader.WG)
	trader.muPairs.Unlock()
	trader.log.Println("STOP: All pairs were interrupted gracefully")
	close(trader.signals)
	close(trader.errors)
	<-trader.stop
	trader.Orders.Shutdown()
	trader.MessageWriters.Shutdown()
}

// AddUser adds user to UserRepository.
func (trader *AlgoTrader) AddUser(user domain.User) error {
	if err := trader.Users.AddUser(&user); err != nil {
		return fmt.Errorf("trader add user error <%w>", err)
	}
	trader.log.Printf("ADD: user: <%s> was added", user.Username)
	return nil
}

// AddPair adds pair and run it if it doesn't exist, else just sign user.
func (trader *AlgoTrader) AddPair(username string, config domain.Config) error {
	if err := config.Validate(); err != nil {
		return err
	}
	user, err := trader.Users.GetUser(username)
	if err != nil {
		return fmt.Errorf("can't add pair in trader: <%w>", err)
	}
	if err = user.AddLimit(config.PairName, config.Limit); err != nil {
		return fmt.Errorf("can't add pair in trader: <%w>", err)
	}
	trader.muPairs.Lock()
	defer trader.muPairs.Unlock()
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
	return nil
}

// AddAndRunPair adds and runs pair.
func (trader *AlgoTrader) AddAndRunPair(pairName string, pairInterval domain.CandleInterval,
	indicatorName string, user *domain.User) error {
	pair, err := NewPair(pairName, pairInterval, indicatorName, trader.log)
	if err != nil {
		return fmt.Errorf("can't create pair in trader: <%w>", err)
	}
	if err = trader.Pairs.AddPair(pair, user); err != nil {
		return fmt.Errorf("can't add pair in trader: <%w>", err)
	}
	trader.log.Printf("ADD: user <%s> was added and pair <%s> <%s> was created",
		user.Username, pair.Name, pair.Interval)
	if err = trader.RunPair(pair); err != nil {
		return fmt.Errorf("can't run pair in trader: <%w>", err)
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
			return fmt.Errorf("can't connect pair in trader: <%w>", err)
		}
	}
	trader.WG.Add(1)
	return nil
}

// DeletePair deletes pair if it is possible.
func (trader AlgoTrader) DeletePair(username string, config domain.Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("can't validate config to delete pair in trader: <%w>", err)
	}
	user, err := trader.Users.GetUser(username)
	if err != nil {
		return err
	}
	trader.muPairs.Lock()
	defer trader.muPairs.Unlock()
	if intervals, ok := trader.Pairs[config.PairName]; ok {
		if pair, ok := intervals[config.PairInterval]; ok {
			if err = pair.DeleteUser(user); err != nil {
				return fmt.Errorf("can't delete user in trader: <%w>", err)
			}
			// pair stopping
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
	return nil
}

// AddMessageWriter adds message writer into slice.
func (trader *AlgoTrader) AddMessageWriter(messageWriter MessageWriter) {
	trader.MessageWriters.AddWriter(messageWriter)
}
