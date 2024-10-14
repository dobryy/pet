package betting

import (
	"github.com/prometheus/client_golang/prometheus"
)

var Collectors = []prometheus.Collector{
	BettingPublishAckDuration,
	BettingPublishAsyncDuration,
	BettingBetDuration,
}

const subsystem = "app_betting"

var defaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5}

var BettingPublishAckDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "publish_ack_duration_seconds",
		Help:      "Betting publish ack duration",
		Buckets:   defaultBuckets,
	},
)

var BettingPublishAsyncDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "publish_async_duration_seconds",
		Help:      "Betting publish async duration",
		Buckets:   defaultBuckets,
	},
)

var BettingBetDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "bet_duration_seconds",
		Help:      "Betting bet duration",
		Buckets:   defaultBuckets,
	},
)
