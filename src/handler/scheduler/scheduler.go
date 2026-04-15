package scheduler

import (
	"sync"

	cfg "go-far/src/config/scheduler"
	"go-far/src/service"

	"github.com/rs/zerolog"
)

var onceSchedulerHandler = &sync.Once{}

type schedulerHandler struct {
	log  zerolog.Logger
	sch  *cfg.Scheduler
	svc  *service.Service
	jobs cfg.SchedulerJobsOptions
}

func InitSchedulerHandler(log zerolog.Logger, sch *cfg.Scheduler, svc *service.Service, jobs cfg.SchedulerJobsOptions) {
	var s *schedulerHandler

	onceSchedulerHandler.Do(func() {
		s = &schedulerHandler{
			log:  log,
			sch:  sch,
			svc:  svc,
			jobs: jobs,
		}

		s.Serve()
	})
}

func (s *schedulerHandler) Serve() *cfg.Scheduler {
	if s.sch == nil {
		s.log.Debug().Msg("Scheduler is disabled, skipping")
		return nil
	}

	// User Generator
	if s.jobs.UserGeneratorJob.Enabled {
		userJob := InitUserGeneratorJob(s.log, s.svc.User, s.jobs.UserGeneratorJob)
		if err := s.sch.AddJob(userJob); err != nil {
			s.log.Error().Err(err).Msg("Failed to add UserGeneratorJob to scheduler")
		}
	}

	// Car Generator
	if s.jobs.CarGeneratorJob.Enabled {
		carJob := InitCarGeneratorJob(s.log, s.svc.Car, s.svc.User, s.jobs.CarGeneratorJob)
		if err := s.sch.AddJob(carJob); err != nil {
			s.log.Error().Err(err).Msg("Failed to add CarGeneratorJob to scheduler")
		}
	}

	// Start scheduler
	s.sch.Start()

	return s.sch
}
