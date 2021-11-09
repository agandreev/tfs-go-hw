package inicators

import (
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
)

const (
	Buy Signal = iota
	Sell
	WaitToSell
	WaitToBuy
)

type Signal int

type Indicator interface {
	Add(candle domain.Candle) (Signal, error)
}
