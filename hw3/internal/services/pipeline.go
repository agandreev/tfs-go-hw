package services

import (
	"fmt"

	"github.com/agandreev/tfs-go-hw/hw3/internal/domain"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// Pipeline represents Candle Pipeline steps with loging.
type Pipeline struct {
	Logger *logrus.Logger
}

// CandlesFromPrices represents second step of pipeline.
// It converts channel of Price to channel of domain.CandlePeriod Candles.
// It returns channel of domain.CandlePeriod Candles and channel of all candles.
func (pipe Pipeline) CandlesFromPrices(g *errgroup.Group, prices <-chan domain.Price,
	period domain.CandlePeriod) <-chan domain.Candle {
	// channel of domain.CandlePeriod candles
	candles := make(chan domain.Candle)
	pipe.Logger.Println(fmt.Sprintf("%s channel is opened", period))

	g.Go(func() error {
		defer pipe.Logger.Println(fmt.Sprintf("%s channel was closed", period))
		defer close(candles)

		gen := domain.CandleGenerator{Period: period,
			Candles: map[string]*domain.Candle{},
			Logger:  pipe.Logger}
		for price := range prices {
			// error stops pipeline
			if err := gen.AddPrice(candles, price); err != nil {
				return err
			}
		}
		gen.RemainCandles(candles)
		return nil
	})

	return candles
}

// CandlesFromCandles represents middle steps of pipeline.
// It converts channel of domain.CandlePeriod from one CandlePeriod to
// channel of domain.CandlePeriod Candles from another CandlePeriod.
// It returns channel of domain.CandlePeriod Candles and channel of all candles.
func (pipe Pipeline) CandlesFromCandles(g *errgroup.Group, lowerCandles <-chan domain.Candle,
	period domain.CandlePeriod) <-chan domain.Candle {
	upperCandles := make(chan domain.Candle)
	pipe.Logger.Println(fmt.Sprintf("%s channel is opened", period))

	g.Go(func() error {
		defer pipe.Logger.Println(fmt.Sprintf("%s channel was closed", period))
		defer close(upperCandles)

		gen := domain.CandleGenerator{Period: period,
			Candles: map[string]*domain.Candle{},
			Logger:  pipe.Logger}
		for candle := range lowerCandles {
			if err := gen.AddCandle(upperCandles, candle); err != nil {
				return err
			}
		}
		gen.RemainCandles(upperCandles)
		return nil
	})

	return upperCandles
}

// SaveCandles saves candles from Candle chan and send them on if isOut is true
func (pipe Pipeline) SaveCandles(g *errgroup.Group,
	candles <-chan domain.Candle, isOut bool) <-chan domain.Candle {
	savedCandles := make(chan domain.Candle)
	pipe.Logger.Println("saving channel is opened")

	g.Go(func() error {
		defer pipe.Logger.Println("saving channel was closed")
		defer close(savedCandles)

		for candle := range candles {
			if isOut {
				savedCandles <- candle
			}
			if err := candle.Save(); err != nil {
				return err
			}
		}
		return nil
	})
	return savedCandles
}
