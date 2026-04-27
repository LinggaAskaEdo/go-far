package pyroscope

import (
	"os"
	"runtime"

	app "go-far/internal/infra/grace"

	"github.com/grafana/pyroscope-go"
	"github.com/rs/zerolog"
)

type PyroscopeOptions struct {
	Enabled bool `yaml:"enabled"`
}

type pyro struct {
	log      *zerolog.Logger
	profiler *pyroscope.Profiler
}

// InitPyroscope initializes the Pyroscope profiler.
// Returns nil if profiling is disabled or fails to start.
// Caller must call Stop() on the returned instance to flush final data.
func InitPyroscope(logger *zerolog.Logger, opt *app.AppOptions) *pyro {
	// Only enable in development/staging environments
	if opt.Environment != "development" && opt.Environment != "staging" {
		logger.Info().Str("env", opt.Environment).Msg("Pyroscope profiling disabled")
		return &pyro{profiler: nil}
	}

	// Extract profiler setup to reduce nesting complexity
	profiler, err := startPyroscope(opt)
	if err != nil {
		logger.Warn().Err(err).Msg("⚠️ Pyroscope failed to start")
		return &pyro{profiler: nil}
	}

	logger.Info().
		Str("app", "go-far-app").
		Str("server", "http://localhost:4040").
		Msg("✅ Pyroscope profiler started")

	// Enable mutex/block profiling (required for those profile types)
	runtime.SetMutexProfileFraction(5) // Sample 1/5 of mutex events
	runtime.SetBlockProfileRate(5)     // Sample 1/5 of blocking events

	return &pyro{
		log:      logger,
		profiler: profiler,
	}
}

// startPyroscope extracts the profiler initialization to reduce function complexity
func startPyroscope(opt *app.AppOptions) (*pyroscope.Profiler, error) {
	return pyroscope.Start(pyroscope.Config{
		ApplicationName: "go-far-app",
		ServerAddress:   "http://localhost:4040",
		Tags: map[string]string{
			"env":      opt.Environment,
			"version":  opt.Version,
			"hostname": os.Getenv("HOSTNAME"),
		},
		Logger: nil,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
}

// Stop gracefully shuts down the profiler and flushes pending data.
// Safe to call even if profiler was never started (nil-safe).
func (p *pyro) Stop() {
	if p.profiler == nil {
		return
	}

	if err := p.profiler.Stop(); err != nil {
		p.log.Error().Err(err).Msg("Failed to stop Pyroscope Profiler")
	}
}
