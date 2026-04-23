package redis

import (
	"context"
	"fmt"
	"time"

	"go-far/internal/preference"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// RedisOptions holds Redis configuration
type RedisOptions struct {
	Password        string        `yaml:"password"`
	Network         string        `yaml:"network"`
	Address         string        `yaml:"address"`
	DialTimeout     time.Duration `yaml:"dial_timeout"`
	CacheTTL        time.Duration `yaml:"cache_ttl"`
	MaxRetries      int           `yaml:"max_retries"`
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	PoolSize        int           `yaml:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxActiveConns  int           `yaml:"max_active_conns"`
	PoolTimeout     time.Duration `yaml:"pool_timeout"`
	Enabled         bool          `yaml:"enabled"`
}

// InitRedis initializes a Redis client
func InitRedis(log *zerolog.Logger, opt *RedisOptions, redisType string) *redis.Client {
	var redisClient *redis.Client
	var DB int

	if !opt.Enabled {
		return nil
	}

	switch redisType {
	case preference.REDIS_APPS:
		DB = 0
	case preference.REDIS_AUTH:
		DB = 11
	default:
		DB = 13
	}

	redisClient = redis.NewClient(&redis.Options{
		Network:         opt.Network,
		Addr:            opt.Address,
		Password:        opt.Password,
		DB:              DB,
		MaxRetries:      opt.MaxRetries,
		MinRetryBackoff: opt.MinRetryBackoff,
		MaxRetryBackoff: opt.MaxRetryBackoff,
		DialTimeout:     opt.DialTimeout,
		ReadTimeout:     opt.ReadTimeout,
		WriteTimeout:    opt.WriteTimeout,
		PoolSize:        opt.PoolSize,
		MinIdleConns:    opt.MinIdleConns,
		MaxIdleConns:    opt.MaxIdleConns,
		MaxActiveConns:  opt.MaxActiveConns,
		PoolTimeout:     opt.PoolTimeout,
	})

	ping, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Panic().Err(err).Msg(fmt.Sprintf("REDIS %s status: %s", redisType, ping))
	}

	log.Info().Msg(fmt.Sprintf("REDIS %s status: %s", redisType, ping))

	return redisClient
}
