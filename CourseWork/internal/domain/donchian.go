package domain

import "fmt"

const (
	DefaultChannelSize = 5
)

// Donchian is implementation of Indicator.
type Donchian struct {
	CandleQueue CandleQueue
	High        float64
	Low         float64
	Medium      float64
	ChannelSize int32
	IsEntered   bool
}

// NewDonchian return pointer to Donchian structure with channel size.
// Channel size could be default.
func NewDonchian() *Donchian {
	donchian := &Donchian{
		CandleQueue: *NewCandleQueue(),
		High:        0,
		Low:         0,
		Medium:      0,
		ChannelSize: DefaultChannelSize,
		IsEntered:   false,
	}
	return donchian
}

// Add is adding implementation of Indicator.
// It adds candle to CandleQueue and process channel's movements.
func (indicator *Donchian) Add(candle Candle) (Signal, error) {
	if !candle.Check() {
		return "", fmt.Errorf("incorrect candle parameters")
	}
	if indicator.CandleQueue.Len() != indicator.ChannelSize {
		indicator.CandleQueue.Add(candle)
		return WaitToSet, nil
	}
	defer func() {
		_ = indicator.CandleQueue.Pop()
		indicator.CandleQueue.Add(candle)
	}()
	indicator.Low = indicator.CandleQueue.Low()
	indicator.High = indicator.CandleQueue.High()
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
