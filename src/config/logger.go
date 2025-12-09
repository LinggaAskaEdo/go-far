package config

import (
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerOptions struct {
	Enabled    bool   `yaml:"enabled"`
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

var onceLogger = sync.Once{}

func InitLogger(opt LoggerOptions) zerolog.Logger {
	var log zerolog.Logger

	onceLogger.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339

		logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
		if err != nil {
			logLevel = int(zerolog.DebugLevel)
		}

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		if opt.Enabled {
			fileLogger := &lumberjack.Logger{
				Filename:   opt.Path,
				MaxSize:    opt.MaxSize,
				MaxBackups: opt.MaxBackups,
				MaxAge:     opt.MaxAge,
				Compress:   opt.Compress,
			}

			output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
		}

		log = zerolog.New(output).
			Level(zerolog.Level(logLevel)).
			With().
			Timestamp().
			Caller().
			Logger()
	})

	return log
}
