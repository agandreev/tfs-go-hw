package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agandreev/tfs-go-hw/hw3/internal/domain"
	"github.com/agandreev/tfs-go-hw/hw3/internal/services"
	"github.com/agandreev/tfs-go-hw/hw3/internal/services/generator"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var tickers = []string{"AAPL", "SBER", "NVDA", "TSLA"}

// init removes files from previous sessions
func init() {
	for _, path := range domain.PeriodFiles {
		err := os.Remove(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			panic(err)
		}
	}
}

func main() {
	logger := log.New()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	runPipeline(ctx, g, logger)

	// signals checking
	g.Go(func() error {
		term := make(chan os.Signal, 1)
		signal.Notify(term, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

		select {
		case <-term:
			log.Println("interruption signal was caught")
			cancel()
		case <-ctx.Done():
			log.Println("stop signal-catcher")
			return ctx.Err()
		}
		return nil
	})

	// pipeline finish checking
	if err := g.Wait(); err == nil || err == context.Canceled {
		log.Println("finished gracefully by interruption")
	} else {
		log.Printf("received error: %v", err)
	}
}

// runPipeline runs all Pipeline's steps
func runPipeline(ctx context.Context, g *errgroup.Group, logger *log.Logger) {
	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  10,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})

	logger.Info("start prices generator...")
	prices := pg.Prices(ctx, g)
	pipe := services.Pipeline{Logger: logger}

	// first step
	firstCandles := pipe.CandlesFromPrices(
		g, prices, domain.CandlePeriod1m)
	firstCandlesSaved := pipe.SaveCandles(g, firstCandles, true)
	// second step
	secondCandles := pipe.CandlesFromCandles(
		g, firstCandlesSaved, domain.CandlePeriod2m)
	secondCandlesSaved := pipe.SaveCandles(g, secondCandles, true)
	// third step
	thirdCandles := pipe.CandlesFromCandles(
		g, secondCandlesSaved, domain.CandlePeriod10m)
	_ = pipe.SaveCandles(g, thirdCandles, false)
}
