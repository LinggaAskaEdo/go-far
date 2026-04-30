package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type PgxPoolCollector struct {
	pool *pgxpool.Pool

	acquireTotal  *prometheus.Desc
	totalConns    *prometheus.Desc
	idleConns     *prometheus.Desc
	acquiredConns *prometheus.Desc
	maxConns      *prometheus.Desc
	waitCount     *prometheus.Desc
	waitDuration  *prometheus.Desc
}

func NewPgxPoolCollector(pool *pgxpool.Pool) *PgxPoolCollector {
	const ns, sub = "pgx", "pool"
	labels := prometheus.Labels{}

	return &PgxPoolCollector{
		pool: pool,
		acquireTotal: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "acquire_total"),
			"Total number of successful connection acquires",
			nil, labels,
		),
		totalConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "total_connections"),
			"Total open connections (idle + acquired)",
			nil, labels,
		),
		idleConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "idle_connections"),
			"Connections sitting idle in the pool",
			nil, labels,
		),
		acquiredConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "acquired_connections"),
			"Connections currently checked out by the app",
			nil, labels,
		),
		maxConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "max_connections"),
			"MaxConns configured on the pool",
			nil, labels,
		),
		waitCount: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "wait_total"),
			"Cumulative number of times a caller waited for a connection",
			nil, labels,
		),
		waitDuration: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "wait_duration_seconds_total"),
			"Cumulative time spent waiting for a connection",
			nil, labels,
		),
	}
}

func (c *PgxPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalConns
	ch <- c.idleConns
	ch <- c.acquiredConns
	ch <- c.maxConns
	ch <- c.waitCount
	ch <- c.waitDuration
}

func (c *PgxPoolCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.pool.Stat()

	ch <- prometheus.MustNewConstMetric(c.acquireTotal, prometheus.CounterValue, float64(stats.AcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue, float64(stats.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(stats.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.acquiredConns, prometheus.GaugeValue, float64(stats.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue, float64(stats.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.waitCount, prometheus.CounterValue, float64(stats.EmptyAcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.waitDuration, prometheus.CounterValue, stats.AcquireDuration().Seconds())
}
