package domain

import (
	"fmt"
	"math"
)

// ErrorJSON is needed to notify client about error through http.
type ErrorJSON struct {
	Message string `json:"error"`
}

// CandleQueue represents candle's sequence for indicators.
type CandleQueue struct {
	Candles []Candle
}

// NewCandleQueue return pointer to CandleQueue.
func NewCandleQueue() *CandleQueue {
	return &CandleQueue{
		Candles: make([]Candle, 0),
	}
}

// Len returns length of CandleQueue.
func (queue CandleQueue) Len() int32 {
	return int32(len(queue.Candles))
}

// Add adds candle to CandleQueue.
func (queue *CandleQueue) Add(candle Candle) {
	queue.Candles = append(queue.Candles, candle)
}

// Pop returns last Candle from CandleQueue and remove it from CandleQueue.
func (queue *CandleQueue) Pop() Candle {
	if len(queue.Candles) == 0 {
		return Candle{}
	}
	candle := queue.Candles[0]
	queue.Candles = queue.Candles[1:]
	return candle
}

// High return the highest Candle values in the CandleQueue.
func (queue CandleQueue) High() float64 {
	var high float64
	for _, candle := range queue.Candles {
		if candle.High > high {
			high = candle.High
		}
	}
	return high
}

// Low returns the lowest Candle in CandleQueue.
func (queue CandleQueue) Low() float64 {
	low := math.MaxFloat64
	for _, candle := range queue.Candles {
		if candle.Low < low {
			low = candle.Low
		}
	}
	return low
}

// StockMarketEvent inform service about necessary order's information.
type StockMarketEvent struct {
	Signal   Signal
	Name     string
	Interval CandleInterval
	Volume   int64
	Close    float64
}

// OrderResponse describes incomplete JSON response.
type OrderResponse struct {
	Result     string      `json:"result"`
	SendStatus *SendStatus `json:"sendStatus"`
}

// SendStatus is a part of OrderResponse.
type SendStatus struct {
	OrderID     string        `json:"order_id"`
	Status      string        `json:"status"`
	OrderEvents []*OrderEvent `json:"orderEvents"`
}

// OrderEvent is a part of SendStatus.
type OrderEvent struct {
	Price               float64              `json:"price"`
	Amount              float64              `json:"amount"`
	Reason              string               `json:"reason"`
	Type                string               `json:"type"`
	OrderPriorExecution *OrderPriorExecution `json:"orderPriorExecution"`
}

// OrderPriorExecution is a part of OrderEvent.
type OrderPriorExecution struct {
	Side string `json:"side"`
}

// Message structs info about OrderResponse into struct for further processing.
func (orderResponse OrderResponse) Message() (Message, error) {
	if orderResponse.SendStatus == nil {
		return Message{}, fmt.Errorf("incorrect received data: SendStatus is nil")
	}
	if orderResponse.SendStatus.OrderEvents == nil ||
		len(orderResponse.SendStatus.OrderEvents) == 0 {
		return Message{}, fmt.Errorf("incorrect received data: OrderEvents is nil")
	}
	if orderResponse.SendStatus.OrderEvents[0].OrderPriorExecution == nil {
		return Message{}, fmt.Errorf(
			"incorrect received data: OrderPriorExecution is nil")
	}
	return Message{
		Name:    "",
		OrderID: orderResponse.SendStatus.OrderID,
		Price:   orderResponse.SendStatus.OrderEvents[0].Price,
		Amount:  int64(orderResponse.SendStatus.OrderEvents[0].Amount),
		Side:    orderResponse.SendStatus.OrderEvents[0].OrderPriorExecution.Side,
	}, nil
}
