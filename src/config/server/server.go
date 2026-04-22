package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"go-far/src/config/middleware"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	onceServer     = sync.Once{}
	httpServerInst *http.Server
)

// ServerOptions holds HTTP server configuration
type ServerOptions struct {
	Mode            string        `yaml:"mode"`
	Port            int           `yaml:"port"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	MaxBodyBytes    int64         `yaml:"max_body_bytes"`
}

// HttpOptions holds HTTP handler configuration
type HttpOptions struct {
	AppName string `yaml:"app_name"`
}

// InitHttpServer initializes the HTTP server
func InitHttpServer(logger *zerolog.Logger, opt *ServerOptions, handler http.Handler) *http.Server {
	onceServer.Do(func() {
		serverPort := fmt.Sprintf(":%d", opt.Port)

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

// InitHttpMux initializes the HTTP ServeMux
func InitHttpMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve swagger files
	swaggerFS := http.Dir("./etc/docs")
	mux.Handle("GET /swagger/", http.StripPrefix("/swagger/", http.FileServer(swaggerFS)))

	return mux
}

// WrapHandler wraps an http.Handler with the middleware chain for server use
func WrapHandler(mux *http.ServeMux, mw middleware.Middleware, opt HttpOptions, serverOpt ServerOptions) http.Handler {
	var handler http.Handler = mux

	maxBodyBytes := serverOpt.MaxBodyBytes
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

	return handler
}
