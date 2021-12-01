package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCandle_Validate(t *testing.T) {
	candle := Candle{}
	assert.NoError(t, candle.Validate())

	candle.Low = -1
	assert.Error(t, candle.Validate())
	candle.Low = 0

	candle.High = -1
	assert.Error(t, candle.Validate())
	candle.High = 0

	candle.Open = -1
	assert.Error(t, candle.Validate())
	candle.Open = 0

	candle.Close = -1
	assert.Error(t, candle.Validate())
	candle.Close = 0

	candle.Volume = -1
	assert.Error(t, candle.Validate())
	candle.Volume = 0

	candle.Time = -1
	assert.Error(t, candle.Validate())
	candle.Time = 0
}

func TestCandle_UnmarshalJSON(t *testing.T) {
	candleJSON := CandleJSON{
		Open:   "1",
		Close:  "1",
		High:   "1",
		Low:    "1",
		Time:   1,
		Volume: 1,
	}
	tmpCandle := Candle{
		Open:   1,
		Close:  1,
		High:   1,
		Low:    1,
		Time:   1,
		Volume: 1,
	}
	data, err := json.Marshal(candleJSON)
	assert.NoError(t, err)
	var unmarshalledCandle Candle
	err = json.Unmarshal(data, &unmarshalledCandle)
	assert.NoError(t, err)
	assert.True(t, tmpCandle.Close == unmarshalledCandle.Close &&
		tmpCandle.Low == unmarshalledCandle.Low &&
		tmpCandle.High == unmarshalledCandle.High &&
		tmpCandle.Volume == unmarshalledCandle.Volume &&
		tmpCandle.Time == unmarshalledCandle.Time)

	badCandles := []CandleJSON{
		{
			Open:   "z",
			Close:  "1",
			High:   "1",
			Low:    "1",
			Time:   1,
			Volume: 1,
		},
		{
			Open:   "1",
			Close:  "z",
			High:   "1",
			Low:    "1",
			Time:   1,
			Volume: 1,
		},
		{
			Open:   "1",
			Close:  "1",
			High:   "z",
			Low:    "1",
			Time:   1,
			Volume: 1,
		},
		{
			Open:   "1",
			Close:  "1",
			High:   "1",
			Low:    "z",
			Time:   1,
			Volume: 1,
		},
	}

	for _, badCandle := range badCandles {
		data, err = json.Marshal(badCandle)
		assert.NoError(t, err)
		err = json.Unmarshal(data, &unmarshalledCandle)
		assert.Error(t, err)
	}

	data, err = json.Marshal(unmarshalledCandle)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &unmarshalledCandle)
	assert.Error(t, err)
}
