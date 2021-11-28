package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDonchian(t *testing.T) {
	donchian := NewDonchian()
	assert.NotNil(t, donchian.CandleQueue)
}

func TestDonchian_Add(t *testing.T) {
	donchian := NewDonchian()
	donchian.ChannelSize = 2
	var i int32
	for i = 0; i < donchian.ChannelSize; i++ {
		signal, _ := donchian.Add(Candle{})
		assert.Equal(t, signal, WaitToSet)
	}
	signal, _ := donchian.Add(Candle{High: 1, Low: 1})
	assert.Equal(t, signal, Buy)
	assert.Equal(t, donchian.IsEntered, true)

	signal, _ = donchian.Add(Candle{High: 1, Low: 1})
	assert.Equal(t, signal, WaitToSell)
	assert.Equal(t, donchian.IsEntered, true)

	signal, _ = donchian.Add(Candle{High: 1, Low: 0})
	assert.Equal(t, signal, Sell)
	assert.Equal(t, donchian.IsEntered, false)

	signal, _ = donchian.Add(Candle{High: 2, Low: 0})

	assert.Equal(t, donchian.High, 1.0)
	assert.Equal(t, donchian.Low, 0.0)

	_, err := donchian.Add(Candle{High: -1})
	assert.Error(t, err)
}
