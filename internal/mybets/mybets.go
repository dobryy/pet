package mybets

import (
	"context"
	"encoding/json"
	"fmt"

	"pet/internal/mybets/consumer"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
)

type MyBets struct {
	jetstream jetstream.JetStream
	consumer  *consumer.Consumer
	logger    *zerolog.Logger
}

func New(js jetstream.JetStream, c *consumer.Consumer, l *zerolog.Logger) *MyBets {
	return &MyBets{
		jetstream: js,
		consumer:  c,
		logger:    l,
	}
}

func (b *MyBets) Run(ctx context.Context) {
	/* 1. When the betting request has been received from FE via some channel it will be sent to the channel
	 * 2. There should be result channel that will be used by FE to wait for the result using websocket
	 * or any other similar technology
	 * 3. It is oversimplified but in real project I'd use channel with some DTOs and create Entity with some
	   e.g. validation logic in it here and send it */
	err := b.consumer.Consume(ctx)
	if err != nil {
		return
	}

}

func (b *MyBets) bet(bet *Ticket) (jetstream.PubAckFuture, error) {
	msg, err := json.Marshal(bet)
	if err != nil {
		return nil, fmt.Errorf("error marshaling message: %s", err)
	}

	ack, err := b.jetstream.PublishAsync(fmt.Sprintf("tickets.%d", bet.ClientId), msg)
	if err != nil {
		b.logger.Error().Msgf("Errr publishing message %s", err)
		return nil, err
	}

	return ack, nil
}
