package scheduler

import (
	"context"
	"strconv"
	"sync"
	"time"

	metricspkg "go-far/internal/infra/metrics"
	"go-far/internal/infra/middleware"
	"go-far/internal/preference"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
)

// Scheduler manages scheduled jobs
type Scheduler struct {
	log            *zerolog.Logger
	cron           *cron.Cron
	jobs           []Job
	mu             sync.RWMutex
	tracingEnabled bool
	metrics        *metricspkg.SchedulerMetrics
}

// Job defines the interface for scheduled jobs
type Job interface {
	Name() string
	Schedule() string
	Run(ctx context.Context) error
}

// SchedulerOptions holds scheduler configuration
type SchedulerOptions struct {
	SchedulerJobs  SchedulerJobsOptions `yaml:"jobs"`
	Enabled        bool                 `yaml:"enabled"`
	TracingEnabled bool                 `yaml:"-"`
}

// SchedulerJobsOptions holds individual job configurations
type SchedulerJobsOptions struct {
	UserGeneratorJob *UserGeneratorJobOptions `yaml:"user_generator"`
	CarGeneratorJob  *CarGeneratorJobOptions  `yaml:"car_generator"`
}

// UserGeneratorJobOptions holds user generator job configuration
type UserGeneratorJobOptions struct {
	RandomUserURL string `yaml:"random_user_url"`
	Cron          string `yaml:"cron"`
	BatchSize     int    `yaml:"batch_size"`
	MinAge        int    `yaml:"min_age"`
	MaxAge        int    `yaml:"max_age"`
	Enabled       bool   `yaml:"enabled"`
}

// CarGeneratorJobOptions holds car generator job configuration
type CarGeneratorJobOptions struct {
	NHTSAAPIURL string `yaml:"nhtsa_api_url"`
	Cron        string `yaml:"cron"`
	BatchSize   int    `yaml:"batch_size"`
	MinYear     int    `yaml:"min_year"`
	MaxYear     int    `yaml:"max_year"`
	Enabled     bool   `yaml:"enabled"`
}

// InitScheduler initializes the scheduler
func InitScheduler(log *zerolog.Logger, opt *SchedulerOptions, tracingEnabled bool, reg *prometheus.Registry) (*Scheduler, *metricspkg.SchedulerMetrics) {
	metrics := metricspkg.NewSchedulerMetrics(reg, &metricspkg.SchedulerJobsOptions{
		UserGeneratorJob: convertUserJobToMetrics(opt.SchedulerJobs.UserGeneratorJob),
		CarGeneratorJob:  convertCarJobToMetrics(opt.SchedulerJobs.CarGeneratorJob),
	})

	return &Scheduler{
		log:            log,
		cron:           cron.New(cron.WithSeconds()),
		jobs:           make([]Job, 0),
		tracingEnabled: tracingEnabled,
		metrics:        metrics,
	}, metrics
}

func convertUserJobToMetrics(src *UserGeneratorJobOptions) *metricspkg.UserGeneratorJobOptions {
	if src == nil {
		return nil
	}

	return &metricspkg.UserGeneratorJobOptions{Enabled: src.Enabled}
}

func convertCarJobToMetrics(src *CarGeneratorJobOptions) *metricspkg.CarGeneratorJobOptions {
	if src == nil {
		return nil
	}

	return &metricspkg.CarGeneratorJobOptions{Enabled: src.Enabled}
}

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.cron.AddFunc(job.Schedule(), func() {
		start := time.Now()
		traceID := middleware.GenerateTraceID()
		spanID := middleware.GenerateSpanID()

		s.log.Info().
			Str("job", job.Name()).
			Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), traceID).
			Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), spanID).
			Msg("Job started")

		ctx := context.WithValue(context.Background(), preference.CONTEXT_KEY_LOG_TRACE_ID, traceID)
		ctx = context.WithValue(ctx, preference.CONTEXT_KEY_LOG_SPAN_ID, spanID)

		execErr := job.Run(ctx)

		duration := time.Since(start)
		if s.metrics != nil {
			s.metrics.RecordExecution(job.Name(), duration, execErr)
		}

		if execErr != nil {
			s.log.Error().
				Err(execErr).
				Str("job", job.Name()).
				Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), traceID).
				Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), spanID).
				Msg("Job execution failed")
			return
		}

		s.log.Info().
			Str("job", job.Name()).
			Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), traceID).
			Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), spanID).
			Msg("Job completed successfully")
	})
	if err != nil {
		return err
	}

	s.jobs = append(s.jobs, job)
	s.log.Info().Str("job", job.Name()).Str("schedule", job.Schedule()).Msg("Job registered")

	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.cron.Start()
	s.log.Debug().Msg("✅ Scheduler started, jobs registered: " + strconv.Itoa(len(s.jobs)))
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx := s.cron.Stop()
	<-ctx.Done()
	s.log.Debug().Msg("Scheduler stopped...")
}

// ListJobs returns the list of registered job names
func (s *Scheduler) ListJobs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobNames := make([]string, len(s.jobs))
	for i, job := range s.jobs {
		jobNames[i] = job.Name()
	}

	return jobNames
}
