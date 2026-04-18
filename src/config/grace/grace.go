package grace

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-far/src/config/tracer"

	"github.com/rs/zerolog"
)

var onceGrace = &sync.Once{}

// App defines the application interface
type App interface {
	Serve()
}

type app struct {
	log             zerolog.Logger
	httpServer      *http.Server
	tracer          tracer.Tracer
	shutdownTimeout time.Duration
}

// InitGrace initializes graceful shutdown handling
func InitGrace(log zerolog.Logger, httpServer *http.Server, tracer tracer.Tracer, shutdownTimeout time.Duration) App {
	var gs *app

	onceGrace.Do(func() {
		gs = &app{
			log:             log,
			httpServer:      httpServer,
			tracer:          tracer,
			shutdownTimeout: shutdownTimeout,
		}
	})

	return gs
}

func (g *app) Serve() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		g.log.Info().Str("addr", g.httpServer.Addr).Msg("Starting HTTP server")
		if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.log.Error().Err(err).Msg("HTTP server error")
		}
	}()

	<-signalCh
	g.log.Info().Msg("Received shutdown signal, gracefully shutting down...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), g.shutdownTimeout)
	defer cancelShutdown()

	if err := g.httpServer.Shutdown(shutdownCtx); err != nil {
		g.log.Error().Err(err).Msg("HTTP server shutdown error")
	}

	if err := g.tracer.Stop(shutdownCtx); err != nil {
		g.log.Error().Err(err).Msg("Tracer shutdown error")
	}

	wg.Wait()
	g.log.Info().Msg("Shutdown complete")
}
