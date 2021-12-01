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
		return fmt.Errorf("can't unmarshal <%w>", err)
	}

	high, err := strconv.ParseFloat(candleJSON.High, 64)
	if err != nil {
		return fmt.Errorf("can't parse high <%w>", err)
	}
	low, err := strconv.ParseFloat(candleJSON.Low, 64)
	if err != nil {
		return fmt.Errorf("can't parse low <%w>", err)
	}
	open, err := strconv.ParseFloat(candleJSON.Open, 64)
	if err != nil {
		return fmt.Errorf("can't parse open <%w>", err)
	}
	cl, err := strconv.ParseFloat(candleJSON.Close, 64)
	if err != nil {
		return fmt.Errorf("can't parse close <%w>", err)
	}
	candle.High = high
	candle.Low = low
	candle.Open = open
	candle.Close = cl
	candle.Time = candleJSON.Time
	candle.Volume = candleJSON.Volume

	return nil
}

// Validate checks candle fields for positive values.
func (candle Candle) Validate() error {
	if candle.Low < 0 {
		return fmt.Errorf("candle low is less than zero")
	}
	if candle.High < 0 {
		return fmt.Errorf("candle high is less than zero")
	}
	if candle.Time < 0 {
		return fmt.Errorf("candle time is less than zero")
	}
	if candle.Close < 0 {
		return fmt.Errorf("candle close is less than zero")
	}
	if candle.Open < 0 {
		return fmt.Errorf("candle open is less than zero")
	}
	if candle.Volume < 0 {
		return fmt.Errorf("candle volume is less than zero")
	}
	return nil
}

func (candle Candle) String() string {
	return fmt.Sprintf("Open: %.2f; High: %.2f; Low: %.2f; Close: %.2f; Volume: %d;",
		candle.Open, candle.High, candle.Low, candle.Close, candle.Volume)
}
