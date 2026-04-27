package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type PgxPoolCollector struct {
	pool *pgxpool.Pool

	totalConns    *prometheus.Desc
	idleConns     *prometheus.Desc
	acquiredConns *prometheus.Desc
	maxConns      *prometheus.Desc
	waitCount     *prometheus.Desc
	waitDuration  *prometheus.Desc
}

func NewPgxPoolCollector(pool *pgxpool.Pool) *PgxPoolCollector {
	const ns, sub = "pgx", "pool"
	label := prometheus.Labels{}

	return &PgxPoolCollector{
		pool: pool,
		totalConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "total_connections"),
			"Total open connections (idle + acquired)",
			nil, label,
		),
		idleConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "idle_connections"),
			"Connections sitting idle in the pool",
			nil, label,
		),
		acquiredConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "acquired_connections"),
			"Connections currently checked out by the app",
			nil, label,
		),
		maxConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "max_connections"),
			"MaxConns configured on the pool",
			nil, label,
		),
		waitCount: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "wait_total"),
			"Cumulative number of times a caller waited for a connection",
			nil, label,
		),
		waitDuration: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "wait_duration_seconds_total"),
			"Cumulative time spent waiting for a connection",
			nil, label,
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
	s := c.pool.Stat()

	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue, float64(s.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(s.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.acquiredConns, prometheus.GaugeValue, float64(s.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue, float64(s.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.waitCount, prometheus.CounterValue, float64(s.EmptyAcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.waitDuration, prometheus.CounterValue, s.AcquireDuration().Seconds())
}
