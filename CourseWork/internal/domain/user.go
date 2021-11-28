package domain

import (
	"fmt"
)

// User describes user's structure with identifying parameters.
type User struct {
	Username   string `json:"username"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	TelegramID int64  `json:"telegram_id"`
}

// NewUser returns pointer to User structure.
func NewUser(username string) *User {
	return &User{
		Username:   username,
		PublicKey:  "",
		PrivateKey: "",
	}
}

// OrderInfo consists of order's information to notify users about their deals.
type OrderInfo struct {
	Name    string
	OrderID string
	Price   float64
	Amount  int64
	Side    string
}

func (orderInfo OrderInfo) String() string {
	return fmt.Sprintf("Name: <%s>,\n"+
		"OrderID: <%s>,\n"+
		"Price: <%.2f>,\n"+
		"Amount: <%d>,\n"+
		"Side: <%s>", orderInfo.Name, orderInfo.OrderID, orderInfo.Price, orderInfo.Amount, orderInfo.Side)
}

// Config consists of necessary information for trading staring.
type Config struct {
	PairName      string         `json:"pair_name"`
	PairInterval  CandleInterval `json:"pair_interval"`
	IndicatorName string         `json:"indicator_name"`
}

// Check checks Config for correct namings.
func (config Config) Check() bool {
	if config.PairInterval != Candle1m && config.PairInterval != Candle2m &&
		config.PairInterval != Candle5m && config.PairInterval != Candle10m {
		return false
	}
	return true
}
