package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	natspkg "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	prometheuspkg "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"pet/internal/betting"
	"pet/pkg/http"
	"pet/pkg/prometheus"
)

var wg sync.WaitGroup

const testTicketsCount = 1_000
const betsChannelSize = 1_000_000

func main() {
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	zl := zerolog.New(os.Stderr).Level(zerolog.InfoLevel).With().Timestamp().Str("service", "Betting").Logger()
	l := &zl

	ctx, cancel := context.WithCancel(context.Background())

	// Prometheus
	prometheuspkg.MustRegister(prometheus.Collectors...)
	prometheusHandler := promhttp.Handler()
	server := http.Init(l, prometheusHandler)

	js, natsCloseFunc := jetstreamConnection(l, ctx)
	defer natsCloseFunc()

	b := betting.New(js, l)

	betsCh := make(chan *betting.Ticket, betsChannelSize)

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.Run(ctx, betsCh)
	}()

	l.Info().Msg("Running...")

	wg.Add(1)
	go func() {
		perf(ctx, betsCh, l)
		defer wg.Done()
	}()

	<-gracefulShutdown
	l.Info().Msg("Shutting down gracefully")
	if err := server.Shutdown(ctx); err != nil {
		l.Err(err).Send()
	}
	cancel()
	wg.Wait()
}

func perf(ctx context.Context, betsCh chan *betting.Ticket, l *zerolog.Logger) {
	// let everything start
	time.Sleep(60 * time.Second)
	l.Info().Msg("Sending tickets to the channel")
	var k int
	for k < testTicketsCount {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}

		k += 1
		betsCh <- &betting.Ticket{
			ClientId:     rand.Intn(testTicketsCount / 100),
			TicketNumber: 1000000000 + k,
			Amount:       0,
			Odds: &betting.Odds{
				Id:    0,
				Value: 0,
			},
		}
		//l.Debug().Msgf("%d", k)
	}
	l.Info().Msg("All tickets have been sent to channel")

	go func() {
		for {
			if len(betsCh) == 0 {
				l.Info().Msg("All tickets in the channel has been taken for processing")
				return
			}
		}
	}()
}

func jetstreamConnection(l *zerolog.Logger, ctx context.Context) (jetstream.JetStream, func()) {
	natsConnection, err := natspkg.Connect("nats://betting:test@nats1:4222,nats://betting:test@nats2:4222,nats://betting:test@nats3:4222")
	if err != nil {
		l.Fatal().Err(err).Msg("Error connecting to NATS")
	}
	l.Info().Msgf("NATS Connected URL: %s", natsConnection.ConnectedUrlRedacted())

	js, err := jetstream.New(natsConnection, jetstream.WithPublishAsyncMaxPending(1_000_000_000))
	if err != nil {
		l.Fatal().Err(err).Msg("Error returning JetStream")
	}

	//_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
	//	Name:     "tickets",
	//	Subjects: []string{"tickets.>"},
	//})

	//if err != nil {
	//	l.Fatal().Err(err).Msg("Error creating stream")
	//}

	go func() {
		tick := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-tick.C:
				pending := js.PublishAsyncPending()
				l.Info().Msgf("Messages pending pubAck: %d", pending)
				break
			}
		}
	}()

	return js, func() { natsConnection.Close() }
}
