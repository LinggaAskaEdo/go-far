package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"go-far/src/config/middleware"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var (
	onceServer     = sync.Once{}
	httpServerInst *http.Server
)

// ServerOptions holds HTTP server configuration
type ServerOptions struct {
	Port            int           `yaml:"port"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	Mode            string        `yaml:"mode"`
}

// GinOptions holds Gin engine configuration
type GinOptions struct {
	AppName string `yaml:"app_name"`
}

// InitHttpServer initializes the HTTP server
func InitHttpServer(logger zerolog.Logger, opt ServerOptions, engine *gin.Engine) *http.Server {
	onceServer.Do(func() {
		serverPort := fmt.Sprintf(":%d", opt.Port)

		httpServerInst = &http.Server{
			Addr:         serverPort,
			WriteTimeout: opt.WriteTimeout,
			ReadTimeout:  opt.ReadTimeout,
			IdleTimeout:  opt.IdleTimeout,
			Handler:      engine,
		}
	})

	return httpServerInst
}

// InitHttpGin initializes the Gin engine
func InitHttpGin(log zerolog.Logger, mw middleware.Middleware, opt GinOptions) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(otelgin.Middleware(opt.AppName))
	router.Use(mw.Handler())
	router.Use(mw.CORS())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.DefaultModelsExpandDepth(-1)))

	return router
}
