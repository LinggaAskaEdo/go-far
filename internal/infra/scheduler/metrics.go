package scheduler

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SchedulerMetrics struct {
	jobExecutionTotal   *prometheus.CounterVec
	jobFailureTotal     *prometheus.CounterVec
	jobDurationSeconds  *prometheus.HistogramVec
	jobLastRunTimestamp *prometheus.GaugeVec
}

func NewSchedulerMetrics(reg *prometheus.Registry, jobs *SchedulerJobsOptions) *SchedulerMetrics {
	m := &SchedulerMetrics{
		jobExecutionTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "scheduler_job_execution_total",
				Help: "Total number of scheduler job executions",
			},
			[]string{"job"},
		),
		jobFailureTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "scheduler_job_failure_total",
				Help: "Total number of scheduler job failures",
			},
			[]string{"job"},
		),
		jobDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "scheduler_job_duration_seconds",
				Help:    "Scheduler job execution duration in seconds",
				Buckets: []float64{.05, .1, .5, 1, 5, 10, 30, 60, 120},
			},
			[]string{"job"},
		),
		jobLastRunTimestamp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "scheduler_job_last_run_timestamp_seconds",
				Help: "Unix timestamp of the last job execution",
			},
			[]string{"job"},
		),
	}

	reg.MustRegister(
		m.jobExecutionTotal,
		m.jobFailureTotal,
		m.jobDurationSeconds,
		m.jobLastRunTimestamp,
	)

	// Pre-initialize all label combinations to 0
	jobNames := getJobNames(jobs)
	for _, job := range jobNames {
		m.jobExecutionTotal.WithLabelValues(job)
		m.jobFailureTotal.WithLabelValues(job)
		m.jobDurationSeconds.WithLabelValues(job)
		m.jobLastRunTimestamp.WithLabelValues(job)
	}

	return m
}

func getJobNames(jobs *SchedulerJobsOptions) []string {
	jobNames := make([]string, 0, 2)
	if jobs.UserGeneratorJob != nil {
		jobNames = append(jobNames, "user_generator")
	}
	if jobs.CarGeneratorJob != nil {
		jobNames = append(jobNames, "car_generator")
	}
	return jobNames
}

func (m *SchedulerMetrics) RecordExecution(job string, duration time.Duration, err error) {
	m.jobExecutionTotal.WithLabelValues(job).Inc()
	m.jobDurationSeconds.WithLabelValues(job).Observe(duration.Seconds())
	m.jobLastRunTimestamp.WithLabelValues(job).SetToCurrentTime()

	if err != nil {
		m.jobFailureTotal.WithLabelValues(job).Inc()
	}
}
