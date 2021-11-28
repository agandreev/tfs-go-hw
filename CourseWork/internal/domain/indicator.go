package domain

const (
	Buy        Signal = "buy"
	Sell       Signal = "sell"
	WaitToSell Signal = "wait to sell"
	WaitToBuy  Signal = "wait to buy"
	WaitToSet  Signal = "wait to set"
)

// Signal is necessary for order creation.
type Signal string

// Indicator implements strategy of stock market service.
type Indicator interface {
	Add(candle Candle) (Signal, error)
}
