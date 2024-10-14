package betting

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
)

const WORKERS_NUMBER = 100

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

func (b *Betting) Run(ctx context.Context, betsCh chan *Ticket, resultsCh chan *Result) {
	b.logger.Info().Msgf("Starting betting workers")

	for w := 1; w <= WORKERS_NUMBER; w++ {
		go func() {
			for {
				select {
				case bet := <-betsCh:
					/* 1. When the betting request has been received from FE via some channel it will be sent to the channel
					 * 2. It is oversimplified but in real project I'd use channel with some DTOs and create Entity with some
					   e.g. validation logic in it here and send it */
					resultsCh <- b.bet(bet)
					break
				case <-ctx.Done():
					b.shutdown()
					return
				}
			}
		}()
	}

	b.logger.Info().Msg("Running...")

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

func (b *Betting) bet(bet *Ticket) *Result {
	result := &Result{
		Ticket: bet,
	}

	msg, err := json.Marshal(bet)
	if err != nil {
		b.logger.Error().Err(err).Msg("Error marshaling message1")
		result.Result = func() error {
			return err

		}
		return result
	}

	publishTime := time.Now()
	pubAckFuture, err := b.jetstream.PublishAsync(fmt.Sprintf("tickets.%d.%d.bet", bet.ClientId, bet.TicketNumber), msg)
	BettingPublishAsyncDuration.Observe(time.Since(publishTime).Seconds())

	if err != nil {
		b.logger.Error().Err(err).Msg("Error publishing message")
		result.Result = func() error {
			return err
		}

		return result
	}

	bet.Time = time.Now()
	result.Result = func() error {
		select {
		case <-pubAckFuture.Ok():
			b.logger.Debug().Msgf("bet ok: %+v", bet)
			return nil
		case pubErr := <-pubAckFuture.Err():
			b.logger.Error().Err(pubErr).Msgf("bet failded: %+v: %w", bet, pubErr)
			return pubErr
		}
	}

	return result
}
