package metrics

import (
	"expvar"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

var httpRequestDurationBuckets = []float64{
	.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10,
}

type MetricsOptions struct {
	Enabled bool `yaml:"enabled"`
}

type Metrics interface {
	RecordHttpRequestDuration(method, path string, status int, duration time.Duration)
	StartPoolMetricsRecorder(interval time.Duration)
	StopPoolMetricsRecorder()
	HTTPHandler() http.Handler
}

type metricsImpl struct {
	log *zerolog.Logger
	httpDuration *prometheus.HistogramVec
}

var (
	onceMetrics sync.Once
	metricsInst Metrics
)

func InitMetrics(log *zerolog.Logger) Metrics {
	onceMetrics.Do(func() {
		metricsInst = &metricsImpl{
			log: log,
			httpDuration: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "http_request_duration_seconds",
					Help:    "HTTP request duration in seconds",
					Buckets: httpRequestDurationBuckets,
				},
				[]string{"method", "path", "status"},
			),
		}
	})

	return metricsInst
}

func (m *metricsImpl) RecordHttpRequestDuration(method, path string, status int, duration time.Duration) {
	m.httpDuration.WithLabelValues(method, path, strconv.Itoa(status)).Observe(duration.Seconds())
}

func (m *metricsImpl) StartPoolMetricsRecorder(interval time.Duration) {
	m.log.Info().Dur("interval", interval).Msg("Started pool metrics recorder")
}

func (m *metricsImpl) StopPoolMetricsRecorder() {
}

func (m *metricsImpl) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/debug/vars" {
			expvar.Handler().ServeHTTP(w, r)
			return
		}
		promhttp.Handler().ServeHTTP(w, r)
	})
}