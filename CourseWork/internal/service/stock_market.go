package service

import (
	"context"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
)

type StockMarketAPI interface {
	AddOrder(domain.StockMarketEvent, *domain.User) (domain.OrderInfo, error)
}

type StockMarketSocket interface {
	Connect(context.Context) error
	SubscribeCandle(Pair, chan domain.Candle, chan PairError) error
}
