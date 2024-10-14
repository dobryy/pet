package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"pet/internal/mybets"
	"pet/internal/mybets/consumer"
	"pet/pkg/http"

	natspkg "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	prometheuspkg "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

var wg sync.WaitGroup

func main() {
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	zl := zerolog.New(os.Stderr).Level(zerolog.InfoLevel).With().Timestamp().Str("service", "MyBets").Logger()
	l := &zl

	ctx, cancel := context.WithCancel(context.Background())
	// Prometheus
	prometheuspkg.MustRegister(mybets.Collectors...)
	prometheusHandler := promhttp.Handler()
	server := http.Init(l, prometheusHandler)

	js, natsCloseFunc := jetstreamConnection(l, ctx)
	defer natsCloseFunc()

	jsConsumer, err := js.CreateOrUpdateConsumer(ctx, "tickets", jetstream.ConsumerConfig{
		Durable:       "mybets",
		FilterSubject: "tickets.>",
		//MaxRequestBatch: 2_000,
	})

	if err != nil {
		l.Fatal().Err(err).Msg("could not create nats consumer")
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				info, _ := jsConsumer.Info(ctx)
				l.Info().Msgf("Pending consumer messages: %d", info.NumWaiting)
			}
		}
	}()

	tc := consumer.NewConsumer(jsConsumer, consumer.NewMsgHandler(), l)
	mb := mybets.New(js, tc, l)

	wg.Add(1)
	go func() {
		defer wg.Done()
		mb.Run(ctx)
	}()

	l.Info().Msg("Running...")

	<-gracefulShutdown
	l.Info().Msg("Shutting down gracefully...")
	if err := server.Shutdown(ctx); err != nil {
		l.Err(err).Send()
	}
	wg.Wait()
	cancel()
}

func jetstreamConnection(l *zerolog.Logger, ctx context.Context) (jetstream.JetStream, func()) {
	natsConnection, err := natspkg.Connect("nats://mybets:test@nats1:4222,nats://mybets:test@nats2:4222,nats://mybets:test@nats3:4222")
	if err != nil {
		l.Fatal().Err(err).Msg("Error connecting to NATS")
	}
	l.Info().Msgf("NATS Connected URL: %s", natsConnection.ConnectedUrlRedacted())

	js, err := jetstream.New(natsConnection)
	if err != nil {
		l.Fatal().Err(err).Msg("Error returning JetStreams")
	}

	if err != nil {
		l.Fatal().Err(err).Msg("Error creating stream")
	}
	return js, func() { natsConnection.Close() }
}
