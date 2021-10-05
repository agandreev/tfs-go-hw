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

func main() {
	logger := log.New()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  10,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})

	logger.Info("start prices generator...")
	prices := pg.Prices(ctx)
	pipe := services.Pipeline{Logger: logger}

	// first step
	firstCandles, allCandles := pipe.CandlesFromPrices(
		g, prices, domain.CandlePeriod1m)
	// second step
	secondCandles, allCandles := pipe.CandlesFromCandles(
		g, firstCandles, allCandles, domain.CandlePeriod2m)
	// third step
	thirdCandles, allCandles := pipe.CandlesFromCandles(
		g, secondCandles, allCandles, domain.CandlePeriod10m)
	// saving
	pipe.SaveAllCandles(g, thirdCandles, allCandles)

	// signals checking
	go func() {
		term := make(chan os.Signal, 1)
		signal.Notify(term, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

		<-term
		cancel()
	}()

	// pipeline finish checking
	if err := g.Wait(); err == nil || err == context.Canceled {
		log.Println("finished gracefully by interruption")
	} else {
		log.Printf("received error: %v", err)
	}
}
