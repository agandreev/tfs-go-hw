package inicators

import (
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
)

const (
	DefaultChannelSize = 20
)

type Donchian struct {
	High   float64
	Medium float64
	Low    float64
	// ChannelSize defines window by time_interval * n
	ChannelSize   int32
	IsEntered     bool
	candleCounter int32
}

func NewDonchian() *Donchian {
	return &Donchian{
		High:          0,
		Low:           0,
		Medium:        0,
		ChannelSize:   DefaultChannelSize,
		IsEntered:     false,
		candleCounter: 0,
	}
}

func (indicator Donchian) Add(candle domain.Candle) (Signal, error) {
	defer indicator.update(candle)
	if indicator.candleCounter+1 > indicator.ChannelSize {
		//todo: it could be broken when you are wait to sell
		indicator.High = candle.High
		indicator.Low = candle.Low
		indicator.candleCounter = 0
		return WaitToBuy, nil
	}
	if indicator.IsEntered {
		if candle.Low < indicator.Medium {
			return Sell, nil
		}
		return WaitToSell, nil
	}
	if candle.High > indicator.High {
		indicator.High = candle.High
		return Buy, nil
	}
	return WaitToBuy, nil
}

func (indicator Donchian) update(candle domain.Candle) {
	if indicator.Low > candle.Low {
		indicator.Low = candle.Low
	}
	if indicator.High < candle.High {
		indicator.High = candle.High
	}
	indicator.Medium = (indicator.High + indicator.Low) / 2
	indicator.candleCounter++
}
