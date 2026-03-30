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
	// User Generator
	if s.jobs.UserGeneratorJob.Enabled {
		userJob := InitUserGeneratorJob(s.log, s.svc.User, s.jobs.UserGeneratorJob)
		if err := s.sch.AddJob(userJob); err != nil {
			s.log.Error().Err(err).Msg("Failed to add UserGeneratorJob to scheduler")
		}
	}

	// Start scheduler
	s.sch.Start()

	return s.sch
}
