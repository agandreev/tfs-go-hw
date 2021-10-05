package services

import (
	"os"

	"github.com/agandreev/tfs-go-hw/hw3/internal/domain"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Pipeline struct {
	Logger *logrus.Logger
}

// CandlesFromPrices represents second step of pipeline.
// It converts channel of Price to channel of domain.CandlePeriod Candles.
// It returns channel of domain.CandlePeriod Candles and channel of all candles.
func (pipe Pipeline) CandlesFromPrices(g *errgroup.Group, prices <-chan domain.Price,
	period domain.CandlePeriod) (<-chan domain.Candle, chan domain.Candle) {
	// channel of domain.CandlePeriod candles
	candles := make(chan domain.Candle)
	// channel of all period candles
	allCandles := make(chan domain.Candle)

	g.Go(func() error {
		defer close(candles)

		gen := domain.CandleGenerator{Period: period,
			Candles: map[string]*domain.Candle{}}
		for price := range prices {
			// error stops pipeline
			if err := gen.AddPrice(candles, price); err != nil {
				return err
			}
		}
		gen.RemainCandles(candles)
		return nil
	})

	return candles, allCandles
}

// CandlesFromCandles represents middle steps of pipeline.
// It converts channel of domain.CandlePeriod from one CandlePeriod to
// channel of domain.CandlePeriod Candles from another CandlePeriod.
// It returns channel of domain.CandlePeriod Candles and channel of all candles.
func (pipe Pipeline) CandlesFromCandles(g *errgroup.Group, candles1m <-chan domain.Candle,
	allCandles chan domain.Candle, period domain.CandlePeriod) (
	<-chan domain.Candle, chan domain.Candle) {
	// channel of domain.CandlePeriod candles
	candles := make(chan domain.Candle)

	g.Go(func() error {
		defer close(candles)

		gen := domain.CandleGenerator{Period: period,
			Candles: map[string]*domain.Candle{}}
		for candle := range candles1m {
			allCandles <- candle
			if err := gen.AddCandle(candles, candle); err != nil {
				return err
			}
		}
		gen.RemainCandles(candles)
		return nil
	})

	return candles, allCandles
}

// SaveAllCandles represents final steps of pipeline.
// It concatenates all channels of domain.CandlePeriod to one.
// And save all candles
func (pipe Pipeline) SaveAllCandles(g *errgroup.Group, lastCandles <-chan domain.Candle,
	allCandles chan domain.Candle) {
	// send Candles from previous step to allCandles for the future processing
	go func() {
		defer close(allCandles)
		for candle := range lastCandles {
			allCandles <- candle
		}
	}()

	g.Go(func() error {
		// remove files from previous session
		if err := removeAllFiles(); err != nil {
			return err
		}
		// save all candles
		for candle := range allCandles {
			if err := candle.Save(); err != nil {
				return err
			}
		}
		return nil
	})
}

// removeAllFiles removes all files from previous session.
func removeAllFiles() error {
	for _, path := range domain.PeriodFiles {
		err := os.Remove(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
	}
	return nil
}
