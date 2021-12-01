package domain

import (
	"errors"
	"fmt"
)

const (
	DefaultChannelSize        = 2
	Buy                Signal = "buy"
	Sell               Signal = "sell"
	WaitToSell         Signal = "wait to sell"
	WaitToBuy          Signal = "wait to buy"
	WaitToSet          Signal = "wait to set"
)

var (
	ErrSameTimestamp = errors.New("candle with this timestamp is already stored")
)

// Signal is necessary for order creation.
type Signal string

// Donchian is implementation of Indicator.
type Donchian struct {
	CandleQueue   CandleQueue
	High          float64
	Low           float64
	Medium        float64
	ChannelSize   int32
	IsEntered     bool
	lastTimestamp int64
}

// NewDonchian return pointer to Donchian structure with channel size.
// Channel size could be default.
func NewDonchian() *Donchian {
	donchian := &Donchian{
		CandleQueue:   *NewCandleQueue(),
		High:          0,
		Low:           0,
		Medium:        0,
		ChannelSize:   DefaultChannelSize,
		IsEntered:     false,
		lastTimestamp: 0,
	}
	return donchian
}

// Add is adding implementation of Indicator.
// It adds candle to CandleQueue and process channel's movements.
func (indicator *Donchian) Add(candle Candle) (Signal, error) {
	if err := candle.Validate(); err != nil {
		return "", fmt.Errorf("incorrect candle parameters in donchian: %w", err)
	}
	if candle.Time <= indicator.lastTimestamp {
		return "", ErrSameTimestamp
	}
	indicator.lastTimestamp = candle.Time
	if indicator.CandleQueue.Len() != indicator.ChannelSize {
		indicator.CandleQueue.Add(candle)
		return WaitToSet, nil
	}
	defer func() {
		_ = indicator.CandleQueue.Pop()
		indicator.CandleQueue.Add(candle)
	}()
	indicator.Low = indicator.CandleQueue.Low
	indicator.High = indicator.CandleQueue.High
	indicator.Medium = (indicator.High + indicator.Low) / 2

	if indicator.IsEntered {
		if candle.Low < indicator.Low {
			indicator.IsEntered = false
			return Sell, nil
		}
		return WaitToSell, nil
	}
	if candle.High > indicator.High {
		indicator.IsEntered = true
		return Buy, nil
	}
	return WaitToBuy, nil
}
