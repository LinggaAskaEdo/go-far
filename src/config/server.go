package config

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var onceServer = sync.Once{}

type ServerOptions struct {
	Port            int           `yaml:"port"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	Mode            string        `yaml:"mode"`
}

func InitHttpServer(logger zerolog.Logger, opt ServerOptions, gin *gin.Engine) *http.Server {
	var server *http.Server

	onceServer.Do(func() {
		serverPort := fmt.Sprintf(":%d", opt.Port)

		server = &http.Server{
			Addr:         serverPort,
			WriteTimeout: opt.WriteTimeout,
			ReadTimeout:  opt.ReadTimeout,
			IdleTimeout:  opt.IdleTimeout,
			Handler:      gin,
		}
	})

	return server
}
