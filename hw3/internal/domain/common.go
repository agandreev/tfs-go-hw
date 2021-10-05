package domain

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// ErrUnknownPeriod raise, when periods of candles don't match
	ErrUnknownPeriod = errors.New("unknown period")
	// ErrWrongTicker raise, when tickers of candles don't match
	ErrWrongTicker = errors.New("wrong ticker")
	// ErrUnclosedCandle raise, when candle's timestamp is outdated
	ErrUnclosedCandle = errors.New("error time period was ended")
)

// CandlePeriod describes Candle's periods
type CandlePeriod string

const (
	CandlePeriod1m  CandlePeriod = "1m"
	CandlePeriod2m  CandlePeriod = "2m"
	CandlePeriod10m CandlePeriod = "10m"
)

var (
	// PeriodFiles stores CandlePeriod's filepath
	PeriodFiles = map[CandlePeriod]string{
		CandlePeriod1m:  "candles_1m.csv",
		CandlePeriod2m:  "candles_2m.csv",
		CandlePeriod10m: "candles_10m.csv",
	}
)

type Price struct {
	Ticker string
	Value  float64
	TS     time.Time
}

// Candle returns Candle consisted of one Price
func (price Price) Candle() *Candle {
	return &Candle{
		Ticker: price.Ticker,
		Open:   price.Value,
		High:   price.Value,
		Low:    price.Value,
		Close:  price.Value,
		TS:     price.TS,
	}
}

// updateTS updates timestamp and period by new entity
func (candle *Candle) updateTS(period CandlePeriod) error {
	ts, err := PeriodTS(period, candle.TS)
	if err != nil {
		return err
	}
	candle.TS = ts
	candle.Period = period
	return nil
}

func PeriodTS(period CandlePeriod, ts time.Time) (time.Time, error) {
	switch period {
	case CandlePeriod1m:
		return ts.Truncate(time.Minute), nil
	case CandlePeriod2m:
		return ts.Truncate(2 * time.Minute), nil
	case CandlePeriod10m:
		return ts.Truncate(10 * time.Minute), nil
	default:
		return time.Time{}, ErrUnknownPeriod
	}
}

type Candle struct {
	Ticker string
	Period CandlePeriod // Интервал
	Open   float64      // Цена открытия
	High   float64      // Максимальная цена
	Low    float64      // Минимальная цена
	Close  float64      // Цена закрытие
	TS     time.Time    // Время начала интервала
}

// CandleGenerator generates Candles of personal CandlePeriod by tickers.
type CandleGenerator struct {
	// Period represent generator's candle period
	Period CandlePeriod
	// Candles is map of Ticker candles
	Candles map[string]*Candle
	// logger
	logger *logrus.Logger
}

// AddPrice convert Price to single price Candle and run AddCandle
func (generator *CandleGenerator) AddPrice(candles chan<- Candle, price Price) error {
	// transfer Price to Candle
	return generator.AddCandle(candles, *price.Candle())
}

// AddCandle add new candle to generator if timestamp allow
// else generator creates new candle and sends old to channel
func (generator *CandleGenerator) AddCandle(candles chan<- Candle, candle Candle) error {
	// update timeStamp
	if err := candle.updateTS(generator.Period); err != nil {
		return err
	}
	// process adding
	if generator.Candles[candle.Ticker] == nil {
		// new candle case
		generator.Candles[candle.Ticker] = &candle
		return nil
	}
	// existing candle case
	err := generator.Candles[candle.Ticker].addCandle(candle)
	// outdated TS case
	if errors.Is(ErrUnclosedCandle, err) {
		candles <- *generator.Candles[candle.Ticker]
		// generator.logger.Println(generator.Candles[candle.Ticker])
		delete(generator.Candles, candle.Ticker)
		return generator.AddCandle(candles, candle)
	}
	// other errors case
	return err
}

// RemainCandles sends сandles with an unfinished interval
func (generator CandleGenerator) RemainCandles(candles chan<- Candle) {
	for _, candle := range generator.Candles {
		if candle != nil {
			//generator.logger.Println(*candle)
			candles <- *candle
		}
	}
}

// addCandle update Candle by new Candle
func (candle *Candle) addCandle(newCandle Candle) error {
	// tickers checking
	if candle.Ticker != newCandle.Ticker {
		return ErrWrongTicker
	}
	// TS checking
	if candle.TS != newCandle.TS {
		return ErrUnclosedCandle
	}
	// updating
	if candle.Low > newCandle.Low {
		candle.Low = newCandle.Low
	}
	if candle.High < newCandle.High {
		candle.High = newCandle.High
	}
	candle.Close = newCandle.Close
	return nil
}

func (candle Candle) String() string {
	return fmt.Sprintf("%s,%s,%f,%f,%f,%f", candle.Ticker, candle.TS,
		candle.Open, candle.High, candle.Low, candle.Close)
}

// Save saves candle to it's CandlePeriod file
func (candle Candle) Save() error {
	filePath, ok := PeriodFiles[candle.Period]
	if !ok {
		return fmt.Errorf("file path to <%s> doesn't exists", candle.Period)
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(candle.String() + "\n")
	// fmt.Println(candle)
	if err != nil {
		return err
	}
	return nil
}
