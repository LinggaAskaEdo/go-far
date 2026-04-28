package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"go-far/internal/infra/middleware"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type HttpServerOptions struct {
	AppName      string        `yaml:"app_name"`
	Mode         string        `yaml:"mode"`
	Port         int           `yaml:"port"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	MaxBodyBytes int64         `yaml:"max_body_bytes"`
}

var (
	onceServer     = sync.Once{}
	onceMetricsSrv = sync.Once{}
	httpServerInst *http.Server
	metricsServer  *http.Server
	handler        http.Handler
)

func InitHttpServer(logger *zerolog.Logger, opt *HttpServerOptions, mw middleware.Middleware, mux *http.ServeMux) *http.Server {
	onceServer.Do(func() {
		serverPort := fmt.Sprintf(":%d", opt.Port)

		handler = mux

		maxBodyBytes := opt.MaxBodyBytes
		if maxBodyBytes == 0 {
			maxBodyBytes = 1 << 20 // 1MB
		}

		// Apply request body size limit to prevent memory exhaustion attacks
		if maxBodyBytes > 0 {
			bodyLimitedHandler := handler // capture current value, not reference
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
				bodyLimitedHandler.ServeHTTP(w, r)
			})
		}

		handler = mw.CORS()(handler)
		handler = mw.Handler()(handler)
		if opt.AppName != "" {
			handler = otelhttp.NewHandler(handler, opt.AppName)
		}

		httpServerInst = &http.Server{
			Addr:         serverPort,
			WriteTimeout: opt.WriteTimeout,
			ReadTimeout:  opt.ReadTimeout,
			IdleTimeout:  opt.IdleTimeout,
			Handler:      handler,
		}
	})

	return httpServerInst
}

func InitMetricsServer(logger *zerolog.Logger, opt *HttpServerOptions, handler http.Handler) *http.Server {
	onceMetricsSrv.Do(func() {
		serverPort := fmt.Sprintf(":%d", opt.Port)

		maxBodyBytes := opt.MaxBodyBytes
		if maxBodyBytes == 0 {
			maxBodyBytes = 1 << 20 // 1MB
		}

		var wrappedHandler = handler
		if maxBodyBytes > 0 {
			bodyLimitedHandler := wrappedHandler
			wrappedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
				bodyLimitedHandler.ServeHTTP(w, r)
			})
		}

		metricsServer = &http.Server{
			Addr:         serverPort,
			WriteTimeout: opt.WriteTimeout,
			ReadTimeout:  opt.ReadTimeout,
			IdleTimeout:  opt.IdleTimeout,
			Handler:      wrappedHandler,
		}
	})

	return metricsServer
}
