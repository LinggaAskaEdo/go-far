package metrics

import (
	"expvar"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

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
	log          *zerolog.Logger
	pool         *pgxpool.Pool
	redisClients []*redis.Client
	httpDuration *prometheus.HistogramVec
}

var (
	reg                        *prometheus.Registry
	onceMetrics                sync.Once
	metricsInst                Metrics
	httpRequestDurationBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	redisClientNames           = []string{"apps", "auth", "limiter"}
)

func getRedisClientName(index int, total int) string {
	if total > 1 && index < len(redisClientNames) {
		return redisClientNames[index]
	}

	return ""
}

func registerRedisCollectors(reg *prometheus.Registry, log *zerolog.Logger, clients []*redis.Client) {
	for i, client := range clients {
		if client == nil {
			continue
		}

		name := getRedisClientName(i, len(clients))
		if err := reg.Register(NewRedisPoolCollector(client, name)); err != nil {
			log.Warn().Err(err).Msg("Failed to register redis collector")
		}
	}
}

func InitMetrics(log *zerolog.Logger, pool *pgxpool.Pool, redisClients ...*redis.Client) Metrics {
	onceMetrics.Do(func() {
		reg = prometheus.NewRegistry()

		httpDuration := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: httpRequestDurationBuckets,
			},
			[]string{"method", "path", "status"},
		)
		if err := reg.Register(httpDuration); err != nil {
			log.Warn().Err(err).Msg("Failed to register httpDuration metric")
		}

		metricsInst = &metricsImpl{
			log:          log,
			pool:         pool,
			redisClients: redisClients,
			httpDuration: httpDuration,
		}

		if pool != nil {
			if err := reg.Register(NewPgxPoolCollector(pool)); err != nil {
				log.Warn().Err(err).Msg("Failed to register pgx collector")
			}
		}

		registerRedisCollectors(reg, log, redisClients)
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
	// Metrics are recorded inline during request handling; no background goroutines to stop
}

func (m *metricsImpl) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/debug/vars" {
			expvar.Handler().ServeHTTP(w, r)
			return
		}

		if reg != nil {
			promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		} else {
			promhttp.Handler().ServeHTTP(w, r)
		}
	})
}
