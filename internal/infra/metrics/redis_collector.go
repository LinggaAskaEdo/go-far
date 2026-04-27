package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type RedisPoolCollector struct {
	client     *redis.Client
	clientName string

	hits       *prometheus.Desc
	misses     *prometheus.Desc
	timeouts   *prometheus.Desc
	totalConns *prometheus.Desc
	idleConns  *prometheus.Desc
	staleConns *prometheus.Desc
	maxConns   *prometheus.Desc
}

func NewRedisPoolCollector(client *redis.Client, clientName string) *RedisPoolCollector {
	const ns, sub = "redis", "pool"
	labels := prometheus.Labels{"client": clientName}

	return &RedisPoolCollector{
		client:     client,
		clientName: clientName,
		hits: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "hits_total"),
			"Times a connection was reused from the pool",
			nil, labels,
		),
		misses: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "misses_total"),
			"Times a new connection was created",
			nil, labels,
		),
		timeouts: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "timeouts_total"),
			"Times a caller timed out waiting for a connection",
			nil, labels,
		),
		totalConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "total_connections"),
			"Total open connections (idle + in-use)",
			nil, labels,
		),
		idleConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "idle_connections"),
			"Connections sitting idle in the pool",
			nil, labels,
		),
		staleConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "stale_connections_total"),
			"Connections removed as stale",
			nil, labels,
		),
		maxConns: prometheus.NewDesc(
			prometheus.BuildFQName(ns, sub, "max_connections"),
			"Max pool size configured on the client",
			nil, labels,
		),
	}
}

func (c *RedisPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.hits
	ch <- c.misses
	ch <- c.timeouts
	ch <- c.totalConns
	ch <- c.idleConns
	ch <- c.staleConns
	ch <- c.maxConns
}

func (c *RedisPoolCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.client.PoolStats()

	ch <- prometheus.MustNewConstMetric(c.hits, prometheus.CounterValue, float64(s.Hits))
	ch <- prometheus.MustNewConstMetric(c.misses, prometheus.CounterValue, float64(s.Misses))
	ch <- prometheus.MustNewConstMetric(c.timeouts, prometheus.CounterValue, float64(s.Timeouts))
	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue, float64(s.TotalConns))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(s.IdleConns))
	ch <- prometheus.MustNewConstMetric(c.staleConns, prometheus.CounterValue, float64(s.StaleConns))
	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue, float64(c.client.Options().PoolSize))
}
