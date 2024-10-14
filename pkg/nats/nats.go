package nats

import (
	natspkg "github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type Nats struct {
	config     *Config
	connection *natspkg.Conn
	logger     *zerolog.Logger
}

func New(l *zerolog.Logger) *Nats {
	return &Nats{
		logger: l,
	}
}

func (n *Nats) Connect() (*natspkg.Conn, func()) {
	nc, err := natspkg.Connect("nats://scores_consumer:test@localhost:4222")
	if err != nil {
		n.logger.Fatal().Err(err).Msg("Error connecting to NATS")
	}
	n.logger.Info().Msgf("NATS Connected URL: %s", nc.ConnectedUrlRedacted())

	n.connection = nc

	return nc, func() {
		nc.Close()
	}
}
