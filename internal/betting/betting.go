package betting

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
)

type Betting struct {
	jetstream jetstream.JetStream
	logger    *zerolog.Logger
}

func New(js jetstream.JetStream, l *zerolog.Logger) *Betting {
	return &Betting{
		jetstream: js,
		logger:    l,
	}
}

func (b *Betting) Run(ctx context.Context, c chan *Ticket) {
	/* 1. When the betting request has been received from FE via some channel it will be sent to the channel
	 * 2. There should be result channel that will be used by FE to wait for the result using websocket
	 * or any other similar technology
	 * 3. It is oversimplified but in real project I'd use channel with some DTOs and create Entity with some
	   e.g. validation logic in it here and send it */

	b.logger.Info().Msg("starting betting")
	for {
		select {
		case bet := <-c:
			go b.bet(bet)
			break
		case <-ctx.Done():
			b.shutdown()
			return
		}
	}
}

func (b *Betting) shutdown() {
	b.logger.Info().Msg("Shutting down betting...")
	tick := time.NewTicker(1 * time.Second)

	select {
	case <-tick.C:
		pending := b.jetstream.PublishAsyncPending()
		b.logger.Info().Msgf("Messages pending pubAck: %d", pending)
		if pending == 0 {
			b.logger.Info().Msg("No bets to be processed. Shut down.")
			return
		}
		break
	}
}

func (b *Betting) bet(bet *Ticket) {
	betstart := time.Now()

	msg, err := json.Marshal(bet)
	if err != nil {
		b.logger.Error().Err(err).Msg("Error marshaling message")
		return
	}

	pubAckFuture, err := b.jetstream.PublishAsync(fmt.Sprintf("tickets.%d.%d.bet", bet.ClientId, bet.TicketNumber), msg)
	if err != nil {
		b.logger.Error().Err(err).Msg("Error publishing message")
		return
	}

	select {
	case <-pubAckFuture.Ok():
		b.logger.Debug().Msgf("bet ok: %+v", bet)
		b.logger.Debug().Msgf("bet ok: %i, time: %i", bet.ClientId, time.Since(betstart).Milliseconds())
		return
	case pubErr := <-pubAckFuture.Err():
		b.logger.Error().Err(pubErr).Msgf("bet failed: %+v: %w", bet, pubErr)
		return
	}
}
