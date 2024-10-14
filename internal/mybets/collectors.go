package mybets

import (
	"github.com/prometheus/client_golang/prometheus"
)

var Collectors = []prometheus.Collector{}

const subsystem = "app_betting"

var defaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5}
