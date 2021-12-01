package domain

import (
	"fmt"
	"math"
	"sync"
)

// ErrorJSON is needed to notify client about error through http.
type ErrorJSON struct {
	Message string `json:"error"`
}

// CandleQueue represents candle's sequence for indicators.
type CandleQueue struct {
	Candles   []Candle
	muCandles *sync.RWMutex
	High      float64
	Low       float64
}

// NewCandleQueue return pointer to CandleQueue.
func NewCandleQueue() *CandleQueue {
	return &CandleQueue{
		Candles:   make([]Candle, 0),
		muCandles: &sync.RWMutex{},
		Low:       math.MaxFloat64,
		High:      math.SmallestNonzeroFloat64,
	}
}

// Len returns length of CandleQueue.
func (queue CandleQueue) Len() int32 {
	queue.muCandles.RLock()
	defer queue.muCandles.RUnlock()
	return int32(len(queue.Candles))
}

// Add adds candle to CandleQueue.
func (queue *CandleQueue) Add(candle Candle) {
	queue.muCandles.Lock()
	defer queue.muCandles.Unlock()
	queue.Candles = append(queue.Candles, candle)
	if candle.High > queue.High {
		queue.High = candle.High
	} else {
		queue.High = queue.getHigh()
	}
	if candle.Low < queue.Low {
		queue.Low = candle.Low
	} else {
		queue.Low = queue.getLow()
	}
}

// Pop returns last Candle from CandleQueue and remove it from CandleQueue.
func (queue *CandleQueue) Pop() Candle {
	queue.muCandles.Lock()
	defer queue.muCandles.Unlock()
	if len(queue.Candles) == 0 {
		return Candle{}
	}
	candle := queue.Candles[0]
	queue.Candles = queue.Candles[1:]
	return candle
}

// getHigh return the highest Candle values in the CandleQueue.
func (queue CandleQueue) getHigh() float64 {
	var high float64
	for _, candle := range queue.Candles {
		if candle.High > high {
			high = candle.High
		}
	}
	return high
}

// getLow returns the lowest Candle in CandleQueue.
func (queue CandleQueue) getLow() float64 {
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

func (event StockMarketEvent) String() string {
	return fmt.Sprintf("Name: %s,\n "+
		"Side: %s,\n"+
		"Interval: %s,\n"+
		"Volume: %d,\n"+
		"Price: %.2f,\n", event.Name, event.Signal, event.Interval,
		event.Volume, event.Close)
}

// StockMarketEventError contains event's error.
type StockMarketEventError struct {
	ErrorMessage string
	Event        StockMarketEvent
}

func (event StockMarketEventError) String() string {
	return fmt.Sprintf("%sError: %s", event.Event.String(), event.ErrorMessage)
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

// Message returns OrderInfo structs info about OrderResponse into struct for further processing.
func (orderResponse OrderResponse) Message(name string) (OrderInfo, error) {
	if orderResponse.SendStatus == nil {
		return OrderInfo{}, fmt.Errorf("incorrect received data: SendStatus is nil")
	}
	if orderResponse.SendStatus.OrderEvents == nil ||
		len(orderResponse.SendStatus.OrderEvents) == 0 {
		return OrderInfo{}, fmt.Errorf("incorrect received data: OrderEvents is nil")
	}
	if orderResponse.SendStatus.OrderEvents[0].OrderPriorExecution == nil {
		return OrderInfo{}, fmt.Errorf(
			"incorrect received data: OrderPriorExecution is nil")
	}
	return OrderInfo{
		Name:    name,
		OrderID: orderResponse.SendStatus.OrderID,
		Price:   orderResponse.SendStatus.OrderEvents[0].Price,
		Amount:  int64(orderResponse.SendStatus.OrderEvents[0].Amount),
		Side:    orderResponse.SendStatus.OrderEvents[0].OrderPriorExecution.Side,
	}, nil
}
