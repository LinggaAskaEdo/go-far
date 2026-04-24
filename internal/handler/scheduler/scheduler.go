package scheduler

import (
	"net/http"
	"sync"

	cfg "go-far/internal/infra/scheduler"
	"go-far/internal/service"

	"github.com/rs/zerolog"
)

var onceSchedulerHandler = &sync.Once{}

type schedulerHandler struct {
	log            *zerolog.Logger
	sch            *cfg.Scheduler
	svc            *service.Service
	jobs           *cfg.SchedulerJobsOptions
	httpClient     *http.Client
	enabled        bool
	tracingEnabled bool
}

func InitSchedulerHandler(log *zerolog.Logger, sch *cfg.Scheduler, svc *service.Service, jobs *cfg.SchedulerJobsOptions, httpClient *http.Client, enabled, tracingEnabled bool) {
	var s *schedulerHandler

	onceSchedulerHandler.Do(func() {
		s = &schedulerHandler{
			log:            log,
			sch:            sch,
			svc:            svc,
			jobs:           jobs,
			httpClient:     httpClient,
			enabled:        enabled,
			tracingEnabled: tracingEnabled,
		}

		s.Serve()
	})
}

func (s *schedulerHandler) Serve() {
	if !s.enabled || s.sch == nil {
		s.log.Debug().Msg("Scheduler is disabled, skipping")
	}

	// User Generator
	if s.jobs.UserGeneratorJob.Enabled {
		userJob := InitUserGeneratorJob(s.log, s.svc.User, s.jobs.UserGeneratorJob, s.httpClient, s.tracingEnabled)
		if err := s.sch.AddJob(userJob); err != nil {
			s.log.Error().Err(err).Msg("Failed to add UserGeneratorJob to scheduler")
		}
	}

	// Car Generator
	if s.jobs.CarGeneratorJob.Enabled {
		carJob := InitCarGeneratorJob(s.log, s.svc.Car, s.svc.User, s.jobs.CarGeneratorJob, s.httpClient, s.jobs.CarGeneratorJob.NHTSAAPIURL, s.tracingEnabled)
		if err := s.sch.AddJob(carJob); err != nil {
			s.log.Error().Err(err).Msg("Failed to add CarGeneratorJob to scheduler")
		}
	}

	// Start scheduler
	s.sch.Start()
}
