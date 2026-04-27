package metrics

import (
	"context"
	"expvar"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type MetricsOptions struct {
	Enabled bool `yaml:"enabled"`
}

type Metrics interface {
	RecordDBMetrics()
	RecordRedisMetrics(redisClient *redis.Client)
	StartPoolMetricsRecorder(interval time.Duration)
	StopPoolMetricsRecorder()
	HTTPHandler() http.Handler
}

type metricsImpl struct {
	log               *zerolog.Logger
	dbPool            *pgxpool.Pool
	redisClient       *redis.Client
	dbPoolConnections *prometheus.GaugeVec
	dbPoolIdle        *prometheus.GaugeVec
	dbPoolTotal       *prometheus.GaugeVec
	redisConnections  prometheus.Gauge
	redisUsedMemory   prometheus.Gauge
	redisKeySpace     prometheus.Gauge
	stopCh            chan struct{}
}

var (
	onceMetrics sync.Once
	metricsInst Metrics
)

func InitMetrics(log *zerolog.Logger, dbPool *pgxpool.Pool, redisClient *redis.Client) Metrics {
	onceMetrics.Do(func() {
		metricsInst = &metricsImpl{
			log:         log,
			dbPool:      dbPool,
			redisClient: redisClient,
			dbPoolConnections: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "db_pool_connections_acquired",
					Help: "Database pool acquired connections",
				},
				[]string{},
			),
			dbPoolIdle: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "db_pool_connections_idle",
					Help: "Database pool idle connections",
				},
				[]string{},
			),
			dbPoolTotal: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "db_pool_connections_total",
					Help: "Database pool total connections",
				},
				[]string{},
			),
			redisConnections: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "redis_connected_clients",
					Help: "Number of connected Redis clients",
				},
			),
			redisUsedMemory: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "redis_memory_used_bytes",
					Help: "Redis used memory in bytes",
				},
			),
			redisKeySpace: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "redis_keyspace_keys",
					Help: "Total number of Redis keys",
				},
			),
			stopCh: make(chan struct{}),
		}
	})

	return metricsInst
}

func (m *metricsImpl) RecordDBMetrics() {
	if m.dbPool == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopCh:
				return
			case <-ticker.C:
				m.recordDBStats()
			}
		}
	}()
}

func (m *metricsImpl) recordDBStats() {
	if m.dbPool == nil {
		return
	}

	stats := m.dbPool.Stat()
	if stats == nil {
		return
	}

	m.dbPoolConnections.WithLabelValues().Set(float64(stats.AcquiredConns()))
	m.dbPoolIdle.WithLabelValues().Set(float64(stats.IdleConns()))
	m.dbPoolTotal.WithLabelValues().Set(float64(stats.TotalConns()))
}

func (m *metricsImpl) RecordRedisMetrics(redisClient *redis.Client) {
	if redisClient == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopCh:
				return
			case <-ticker.C:
				m.recordRedisStats(redisClient)
			}
		}
	}()
}

func (m *metricsImpl) recordRedisStats(redisClient *redis.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	info, err := redisClient.Info(ctx, "clients,memory,keyspace").Result()
	if err != nil {
		m.log.Error().Err(err).Msg("Failed to get Redis info")
		return
	}

	var connectedClients int64
	var usedMemory int64
	var totalKeys int64

	for line := range strings.SplitSeq(info, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "connected_clients:") {
			connectedClients, _ = strconv.ParseInt(strings.TrimSpace(strings.Split(line, ":")[1]), 10, 64)
		}
		if strings.HasPrefix(line, "used_memory:") {
			usedMemory, _ = strconv.ParseInt(strings.TrimSpace(strings.Split(line, ":")[1]), 10, 64)
		}
		if strings.HasPrefix(line, "db0:") {
			dbLine := strings.TrimSpace(strings.Split(line, ":")[1])
			for part := range strings.SplitSeq(dbLine, ",") {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "keys=") {
					totalKeys, _ = strconv.ParseInt(strings.TrimSpace(strings.Split(part, "=")[1]), 10, 64)
				}
			}
		}
	}

	m.redisConnections.Set(float64(connectedClients))
	m.redisUsedMemory.Set(float64(usedMemory))
	m.redisKeySpace.Set(float64(totalKeys))
}

func (m *metricsImpl) StartPoolMetricsRecorder(interval time.Duration) {
	m.log.Info().Dur("interval", interval).Msg("Started pool metrics recorder")
}

func (m *metricsImpl) StopPoolMetricsRecorder() {
	close(m.stopCh)
	m.stopCh = make(chan struct{})
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
