package http

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func Init(
	l *zerolog.Logger,
	metricsHandler http.Handler,
) *http.Server {
	l.Info().Msg("Initiating HTTP Server")

	h := http.DefaultServeMux
	h.Handle("/metrics", metricsHandler)

	s := http.Server{
		Addr:              ":8080",
		Handler:           h,
		ReadHeaderTimeout: 60 * time.Second,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			l.Err(err).Send()
		}
	}()

	return &s
}
