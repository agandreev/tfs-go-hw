package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCandle_Check(t *testing.T) {
	candle := Candle{}
	assert.True(t, candle.Check())

	candle.Low = -1
	assert.False(t, candle.Check())
	candle.Low = 0

	candle.High = -1
	assert.False(t, candle.Check())
	candle.High = 0

	candle.Open = -1
	assert.False(t, candle.Check())
	candle.Open = 0

	candle.Close = -1
	assert.False(t, candle.Check())
	candle.Close = 0

	candle.Volume = -1
	assert.False(t, candle.Check())
	candle.Volume = 0

	candle.Time = -1
	assert.False(t, candle.Check())
	candle.Time = 0
}
