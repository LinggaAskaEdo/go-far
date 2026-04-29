package grace

import (
	"context"
	"net"
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
	metricsSrv *http.Server
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
func InitGrace(log *zerolog.Logger, httpServer *http.Server, metricsSrv *http.Server, timeout time.Duration) App {
	var gs *app

	onceGrace.Do(func() {
		gs = &app{
			log:        log,
			httpServer: httpServer,
			metricsSrv: metricsSrv,
			timeout:    timeout,
		}
	})

	return gs
}

func (g *app) Serve() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	var wg sync.WaitGroup
	listenConfig := net.ListenConfig{
		KeepAlive: 30 * time.Second,
	}

	// Main HTTP server
	httpListener, err := listenConfig.Listen(context.Background(), "tcp", g.httpServer.Addr)
	if err != nil {
		g.log.Error().Err(err).Str("addr", g.httpServer.Addr).Msg("Failed to listen for HTTP server")
		return
	}

	g.log.Info().Msg("✅ HTTP server already started at " + g.httpServer.Addr)

	wg.Go(func() {
		if serveErr := g.httpServer.Serve(httpListener); serveErr != nil && serveErr != http.ErrServerClosed {
			g.log.Error().Err(serveErr).Msg("HTTP server error")
		}
	})

	// Metrics server (if enabled)
	var metricsListener net.Listener
	if g.metricsSrv != nil {
		metricsListener, err = listenConfig.Listen(context.Background(), "tcp", g.metricsSrv.Addr)
		if err != nil {
			g.log.Error().Err(err).Str("addr", g.metricsSrv.Addr).Msg("Failed to listen for metrics HTTP server")
			return
		}

		g.log.Info().Msg("✅ Metrics HTTP server already started at " + g.metricsSrv.Addr)

		wg.Go(func() {
			if serveErr := g.metricsSrv.Serve(metricsListener); serveErr != nil && serveErr != http.ErrServerClosed {
				g.log.Error().Err(serveErr).Msg("Metrics HTTP server error")
			}
		})
	}

	<-signalCh
	g.log.Debug().Msg("Received shutdown signal, gracefully shutting down...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), g.timeout)
	defer cancelShutdown()

	// Shutdown main server
	if err := g.httpServer.Shutdown(shutdownCtx); err != nil {
		g.log.Error().Err(err).Msg("HTTP server shutdown error")
	}

	// Shutdown metrics server if present and was successfully started
	if g.metricsSrv != nil {
		if err := g.metricsSrv.Shutdown(shutdownCtx); err != nil {
			g.log.Error().Err(err).Msg("Metrics HTTP shutdown error")
		}
	}

	wg.Wait()
	g.log.Debug().Msg("Shutdown complete")
}
