package service

import (
	"context"
	"fmt"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	wsLink = "wss://futures.kraken.com/ws/v1?chart"
)

// KrakenSocket implements socket connections to the stock market.
type KrakenSocket struct {
	ws  *websocket.Conn
	ctx context.Context
	log *logrus.Logger
}

// SubscriptionMessage describes JSON subscription response.
type SubscriptionMessage struct {
	Event      string   `json:"event"`
	Feed       string   `json:"feed"`
	ProductIDs []string `json:"product_ids,omitempty"`
	Version    int64    `json:"version,omitempty"`
}

// CandleMessage describes JSON candle response.
type CandleMessage struct {
	Feed      string        `json:"feed"`
	Candle    domain.Candle `json:"candle"`
	ProductID string        `json:"product_id"`
	Time      int64         `json:"time"`
}

// Connect dials to the socket.
func (socket *KrakenSocket) Connect(ctx context.Context) error {
	connection, response, err := websocket.DefaultDialer.Dial(wsLink, nil)
	if err != nil {
		socket.log.Printf("ERROR creating connection to url, %v", err)
	} else {
		socket.log.Printf("SUCCESS: Connection established with %s  \n", wsLink)

		if err != nil {
			socket.log.Println("Error marshaling json:", err)
		}
		socket.log.Printf("RESPONSE: %+v", *response)
	}
	socket.ws = connection
	socket.ctx = ctx
	return fmt.Errorf("error in socket connection: <%w>", err)
}

// SubscribeCandle subscribes candle's socket.
func (socket KrakenSocket) SubscribeCandle(pair Pair, candles chan domain.Candle,
	errors chan PairError) error {
	subscriptionCandleRequest := SubscriptionMessage{
		Event:      "subscribe",
		Feed:       string(pair.Interval),
		ProductIDs: []string{pair.Name},
	}

	subscriptionHeartbeatRequest := SubscriptionMessage{
		Event: "subscribe",
		Feed:  "heartbeat",
	}

	err := socket.ws.WriteJSON(subscriptionHeartbeatRequest)
	if err != nil {
		return fmt.Errorf("error in heartbeat subscription: <%w>", err)
	}

	err = socket.ws.WriteJSON(subscriptionCandleRequest)
	if err != nil {
		return fmt.Errorf("error in candle subscription: <%w>", err)
	}

	err = socket.readInfoMessages()
	if err != nil {
		return fmt.Errorf("error in subscription reading: <%w>", err)
	}

	socket.runPipeLine(pair, candles, errors)

	return nil
}

// runPipeLine runs goroutine for candle reading.
func (socket KrakenSocket) runPipeLine(pair Pair, candles chan domain.Candle,
	errors chan PairError) {
	go func() {
		defer close(candles)
		candleMessage := CandleMessage{}
		for {
			select {
			case <-socket.ctx.Done():
				return
			default:
				if err := socket.ws.ReadJSON(&candleMessage); err != nil {
					socket.log.Printf("ERROR: <%s> has lost connection <%s>", pair.Name, err)
					errors <- PairError{
						Name:     pair.Name,
						Interval: pair.Interval,
						Message:  err.Error(),
					}
					return
				}
				if !(candleMessage.Feed == "heartbeat") {
					candles <- candleMessage.Candle
				}
			}
		}
	}()
}

// readInfoMessages skips first info messages from socket.
func (socket KrakenSocket) readInfoMessages() error {
	subscriptionResponse := SubscriptionMessage{}
	// wait for {event: "info", version: 1}
	err := socket.ws.ReadJSON(&subscriptionResponse)
	if err != nil {
		return err
	}
	if subscriptionResponse.Event != "info" || subscriptionResponse.Version != 1 {
		return fmt.Errorf("incorrect info response")
	}
	// wait for {event: "subscribed"...}
	err = socket.ws.ReadJSON(&subscriptionResponse)
	if err != nil {
		return err
	}
	if subscriptionResponse.Event != "subscribed" {
		fmt.Println(subscriptionResponse)
		return fmt.Errorf("subscription is failed")
	}
	// wait for {event: "subscribed"...}
	err = socket.ws.ReadJSON(&subscriptionResponse)
	if err != nil {
		return err
	}
	if subscriptionResponse.Event != "subscribed" {
		fmt.Println(subscriptionResponse)
		return fmt.Errorf("subscription is failed")
	}
	return nil
}
