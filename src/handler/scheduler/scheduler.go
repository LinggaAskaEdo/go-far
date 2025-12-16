package scheduler

import (
	"sync"

	"go-far/src/config"
	"go-far/src/service"

	"github.com/rs/zerolog"
)

var onceScheduler = &sync.Once{}

type scheduler struct {
	log  zerolog.Logger
	sch  *config.Scheduler
	svc  *service.Service
	jobs config.SchedulerJobsOptions
}

func InitSchedulerHandler(log zerolog.Logger, sch *config.Scheduler, svc *service.Service, jobs config.SchedulerJobsOptions) {
	var s *scheduler

	onceScheduler.Do(func() {
		s = &scheduler{
			log:  log,
			sch:  sch,
			svc:  svc,
			jobs: jobs,
		}

		s.Serve()
	})
}

func (s *scheduler) Serve() *config.Scheduler {
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
