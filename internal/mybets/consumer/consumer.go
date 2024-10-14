package consumer

import (
	"context"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
)

const nakDelay = 5 * time.Second

type Consumer struct {
	consumer   jetstream.Consumer
	msgHandler *MsgHandler
	logger     *zerolog.Logger
}

func NewConsumer(c jetstream.Consumer, h *MsgHandler, l *zerolog.Logger) *Consumer {
	loggerWithContext := l.With().Str("Consumer", c.CachedInfo().Name).Logger()

	return &Consumer{
		consumer:   c,
		msgHandler: h,
		logger:     &loggerWithContext,
	}
}

func (c *Consumer) Consume(ctx context.Context) error {
	c.logger.Debug().Msgf("Consuming...")
	_, err := c.consumer.Consume(func(msg jetstream.Msg) {
		c.logger.Debug().Msgf("received %s\n", msg.Subject())

		// TODO depending on the subject -> apply changes to the ticket in KV store
		// to make it the way it will be present on FE
		err := c.msgHandler.Handle(ctx, msg)

		if err != nil {
			c.logger.Err(err).Msgf("Failed to handle message %s", msg.Headers().Get("Nats-Msg-Id"))
			c.nak(msg)
			return
		}

		err = msg.Ack()
		if err != nil {
			c.logger.Err(err).Msgf("Failed to ACK message %s", msg.Headers().Get("Nats-Msg-Id"))
			return
		}

		c.logger.Debug().Msgf("ACKed message %s", msg.Headers().Get("Nats-Msg-Id"))

		//}()
	})

	if err != nil {
		c.logger.Fatal().Err(err).Msg("Failed to consume")
	}
	return nil
}

func (c *Consumer) nak(msg jetstream.Msg) {
	err := msg.NakWithDelay(nakDelay)
	if err != nil {
		c.logger.Err(err).Msgf("Failed to NAK message %s with delay", msg.Headers().Get("Nats-Msg-Id"))
		return
	}

	c.logger.Debug().Msgf("Message has been NAKed with a delay of %f seconds", nakDelay.Seconds())
	return
}
