package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

var (
	onceGrace = &sync.Once{}
	wg        sync.WaitGroup
)

type App interface {
	Serve()
}

type app struct {
	log        zerolog.Logger
	httpServer *http.Server
}

func InitGrace(log zerolog.Logger, httpServer *http.Server) App {
	var gs *app

	onceGrace.Do(func() {
		gs = &app{
			log:        log,
			httpServer: httpServer,
		}
	})

	return gs
}

func (g *app) Serve() {
	ctx, cancel := context.WithCancel(context.Background())

	// Listen for termination signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	wg.Add(1)
	go startHTTPServer(ctx, &wg, g.log, g.httpServer)

	// Wait for termination signal
	<-signalCh

	// Start the graceful shutdown process
	g.log.Debug().Msg("Gracefully shutting down HTTP server...")

	// Cancel the context to signal the HTTP server to stop
	cancel()

	// Wait for the HTTP server to finish
	wg.Wait()

	g.log.Debug().Msg("Shutdown complete.")
}

func startHTTPServer(ctx context.Context, wg *sync.WaitGroup, log zerolog.Logger, httpServer *http.Server) {
	defer wg.Done()

	// Start the HTTP server in a separate goroutine
	go func() {
		log.Debug().Msg("Starting HTTP server...")
		log.Debug().Msg(fmt.Sprintf("HTTP server start on %s", httpServer.Addr))

		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error().AnErr("HTTP server error", err)
		}
	}()

	// Wait for the context to be canceled
	<-ctx.Done()
	log.Debug().Msg("HTTP server started...")

	// Shutdown the server gracefully
	log.Debug().Msg("Shutting down HTTP server gracefully...")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	err := httpServer.Shutdown(shutdownCtx)
	if err != nil {
		log.Debug().AnErr("HTTP server shutdown error", err)
	}

	log.Debug().Msg("HTTP server stopped.")
}
