package domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
		signal, err := donchian.Add(Candle{Time: int64(i + 1)})
		assert.NoError(t, err)
		assert.Equal(t, signal, WaitToSet)
	}
	i++
	signal, err := donchian.Add(Candle{High: 1, Low: 1, Time: int64(i)})
	assert.NoError(t, err)
	assert.Equal(t, signal, Buy)
	assert.Equal(t, donchian.IsEntered, true)

	fmt.Println(donchian.Low)
	i++
	signal, err = donchian.Add(Candle{High: 1, Low: 1, Time: int64(i)})
	assert.NoError(t, err)
	assert.Equal(t, signal, WaitToSell)
	assert.Equal(t, donchian.IsEntered, true)

	i++
	fmt.Println(donchian.Low)
	signal, err = donchian.Add(Candle{High: 1, Low: 0, Time: int64(i)})
	fmt.Println(donchian.Low)
	assert.NoError(t, err)
	assert.Equal(t, signal, Sell)
	assert.Equal(t, donchian.IsEntered, false)

	i++
	_, err = donchian.Add(Candle{High: 2, Low: 0, Time: int64(i)})
	assert.NoError(t, err)
	assert.Equal(t, donchian.High, 1.0)
	assert.Equal(t, donchian.Low, 0.0)

	i++
	_, err = donchian.Add(Candle{High: -1, Time: int64(i)})
	assert.Error(t, err)
	_, err = donchian.Add(Candle{High: 3})
	assert.Error(t, err)
}
