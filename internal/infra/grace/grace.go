package grace

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

type App interface {
	Serve()
}

type app struct {
	log        *zerolog.Logger
	httpServer *http.Server
	timeout    time.Duration
}

type AppOptions struct {
	Name            string        `yaml:"name"`
	Version         string        `yaml:"version"`
	Environment     string        `yaml:"environment"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

var onceGrace = &sync.Once{}

// InitGrace initializes graceful shutdown handling
func InitGrace(log *zerolog.Logger, httpServer *http.Server, timeout time.Duration) App {
	var gs *app

	onceGrace.Do(func() {
		gs = &app{
			log:        log,
			httpServer: httpServer,
			timeout:    timeout,
		}
	})

	return gs
}

func (g *app) Serve() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	var wg sync.WaitGroup
	wg.Go(func() {
		g.log.Info().Str("addr", g.httpServer.Addr).Msg("✅ Starting HTTP server")
		if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.log.Error().Err(err).Msg("HTTP server error")
		}
	})

	<-signalCh
	g.log.Debug().Msg("Received shutdown signal, gracefully shutting down...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), g.timeout)
	defer cancelShutdown()

	if err := g.httpServer.Shutdown(shutdownCtx); err != nil {
		g.log.Error().Err(err).Msg("HTTP server shutdown error")
	}

	wg.Wait()
	g.log.Debug().Msg("Shutdown complete")
}
