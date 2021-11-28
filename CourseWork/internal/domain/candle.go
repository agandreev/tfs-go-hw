package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	Candle1m  CandleInterval = "candles_trade_1m"
	Candle2m  CandleInterval = "candles_trade_2m"
	Candle5m  CandleInterval = "candles_trade_5m"
	Candle10m CandleInterval = "candles_trade_10m"
)

// CandleInterval represents string description of candle time interval.
type CandleInterval string

// Candle consists of all possible properties from kraken stock market.
type Candle struct {
	Interval CandleInterval
	Open     float64 `json:"open"`
	Close    float64 `json:"close"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Time     int64   `json:"time"`
	Volume   int64   `json:"volume"`
}

// CandleJSON describes json from stock market with wrong field types.
type CandleJSON struct {
	Open   string `json:"open"`
	Close  string `json:"close"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Time   int64  `json:"time"`
	Volume int64  `json:"volume"`
}

// UnmarshalJSON is custom unmarshaler which reflects CandleJSON to Candle.
func (candle *Candle) UnmarshalJSON(data []byte) error {
	var candleJSON CandleJSON
	if err := json.Unmarshal(data, &candleJSON); err != nil {
		return err
	}

	high, err := strconv.ParseFloat(candleJSON.High, 64)
	if err != nil {
		return err
	}
	low, err := strconv.ParseFloat(candleJSON.Low, 64)
	if err != nil {
		return err
	}
	open, err := strconv.ParseFloat(candleJSON.Open, 64)
	if err != nil {
		return err
	}
	cl, err := strconv.ParseFloat(candleJSON.Close, 64)
	if err != nil {
		return err
	}
	candle.High = high
	candle.Low = low
	candle.Open = open
	candle.Close = cl
	candle.Time = candleJSON.Time
	candle.Volume = candleJSON.Volume

	return nil
}

// Check checks candle fields for positive values.
func (candle Candle) Check() bool {
	if candle.Low < 0 || candle.High < 0 || candle.Time < 0 ||
		candle.Close < 0 || candle.Open < 0 || candle.Volume < 0 {
		return false
	}
	return true
}

func (candle Candle) String() string {
	return fmt.Sprintf("Open: %.2f; High: %.2f; Low: %.2f; Close: %.2f; Volume: %d;",
		candle.Open, candle.High, candle.Low, candle.Close, candle.Volume)
}
