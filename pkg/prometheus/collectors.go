package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

var Collectors = []prometheus.Collector{
	BettingPublishingDuration,
}

const subsystem = "app_betting"

var defaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5}

var BettingPublishingDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "betting_publishing_duration_seconds",
		Help:      "Betting publishing duration",
		Buckets:   defaultBuckets,
	},
	[]string{"worker", "subject"},
)
