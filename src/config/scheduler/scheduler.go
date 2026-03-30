package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
)

// Scheduler manages scheduled jobs
type Scheduler struct {
	log  zerolog.Logger
	cron *cron.Cron
	jobs []Job
	mu   sync.RWMutex
}

// Job defines the interface for scheduled jobs
type Job interface {
	Name() string
	Schedule() string
	Run(ctx context.Context) error
}

// SchedulerOptions holds scheduler configuration
type SchedulerOptions struct {
	Enabled       bool                 `yaml:"enabled"`
	SchedulerJobs SchedulerJobsOptions `yaml:"jobs"`
}

// SchedulerJobsOptions holds individual job configurations
type SchedulerJobsOptions struct {
	UserGeneratorJob UserGeneratorJobOptions `yaml:"user_generator"`
}

// UserGeneratorJobOptions holds user generator job configuration
type UserGeneratorJobOptions struct {
	Enabled   bool   `yaml:"enabled"`
	Cron      string `yaml:"cron"`
	BatchSize int    `yaml:"batch_size"`
	MinAge    int    `yaml:"min_age"`
	MaxAge    int    `yaml:"max_age"`
}

// InitScheduler initializes the scheduler
func InitScheduler(log zerolog.Logger, opt SchedulerOptions) *Scheduler {
	if opt.Enabled {
		return &Scheduler{
			log:  log,
			cron: cron.New(cron.WithSeconds()),
			jobs: make([]Job, 0),
		}
	}

	return nil
}

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.cron.AddFunc(job.Schedule(), func() {
		s.log.Info().Str("job", job.Name()).Msg("Job started")

		ctx := context.Background()
		if err := job.Run(ctx); err != nil {
			s.log.Error().Err(err).Str("job", job.Name()).Msg("Job execution failed")
			return
		}

		s.log.Info().Str("job", job.Name()).Msg("Job completed successfully")
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
	s.log.Debug().Msg("Scheduler started, jobs registered: " + fmt.Sprint(len(s.jobs)))
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
