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

var (
	onceGrace = &sync.Once{}
	wg        sync.WaitGroup
)

// App defines the application interface
type App interface {
	Serve()
}

type app struct {
	log        zerolog.Logger
	httpServer *http.Server
	tracer     tracer.Tracer
}

// InitGrace initializes graceful shutdown handling
func InitGrace(log zerolog.Logger, httpServer *http.Server, tracer tracer.Tracer) App {
	var gs *app

	onceGrace.Do(func() {
		gs = &app{
			log:        log,
			httpServer: httpServer,
			tracer:     tracer,
		}
	})

	return gs
}

func (g *app) Serve() {
	ctx, cancel := context.WithCancel(context.Background())

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	wg.Add(1)
	go startHTTPServer(ctx, &wg, g.log, g.httpServer, g.tracer)

	<-signalCh

	g.log.Debug().Msg("Gracefully shutting down HTTP server...")
	cancel()
	wg.Wait()
	g.log.Debug().Msg("Shutdown complete...")
}

func startHTTPServer(ctx context.Context, wg *sync.WaitGroup, log zerolog.Logger, httpServer *http.Server, tracer tracer.Tracer) {
	defer wg.Done()

	go func() {
		log.Debug().Msg("Starting HTTP server...")
		log.Debug().Msg("HTTP server start on " + httpServer.Addr)

		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error().AnErr("HTTP server error", err)
		}
	}()

	<-ctx.Done()
	log.Debug().Msg("HTTP server started...")

	log.Debug().Msg("Shutting down HTTP server gracefully...")
	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, 5*time.Second)
	defer cancelShutdown()

	err := httpServer.Shutdown(shutdownCtx)
	if err != nil {
		log.Debug().AnErr("HTTP server shutdown error", err)
	}

	err = tracer.Stop(shutdownCtx)
	if err != nil {
		log.Debug().AnErr("Tracer shutdown error", err)
	}

	log.Debug().Msg("HTTP server stopped...")
}
