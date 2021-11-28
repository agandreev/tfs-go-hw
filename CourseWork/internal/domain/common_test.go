package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCandleQueue(t *testing.T) {
	candleQueue := NewCandleQueue()
	assert.Equal(t, len(candleQueue.Candles), 0)
}

func TestCandleQueue_Add(t *testing.T) {
	candleQueue := NewCandleQueue()
	candleQueue.Add(Candle{})
	assert.Equal(t, len(candleQueue.Candles), 1)
}

func TestCandleQueue_Pop(t *testing.T) {
	candleQueue := NewCandleQueue()
	candleQueue.Add(Candle{})
	assert.Equal(t, candleQueue.Pop(), Candle{})
	assert.Equal(t, len(candleQueue.Candles), 0)
}

func TestCandleQueue_High(t *testing.T) {
	candleQueue := NewCandleQueue()
	candleQueue.Add(Candle{High: 1})
	candleQueue.Add(Candle{High: 3})
	candleQueue.Add(Candle{High: 2})
	assert.Equal(t, candleQueue.High(), 3.0)
}

func TestCandleQueue_Low(t *testing.T) {
	candleQueue := NewCandleQueue()
	candleQueue.Add(Candle{Low: 3})
	candleQueue.Add(Candle{Low: 1})
	candleQueue.Add(Candle{Low: 2})
	assert.Equal(t, candleQueue.Low(), 1.0)
}

func TestCandleQueue_Len(t *testing.T) {
	candleQueue := NewCandleQueue()
	assert.Equal(t, candleQueue.Len(), int32(0))
	candleQueue.Add(Candle{})
	assert.Equal(t, candleQueue.Len(), int32(1))
}
