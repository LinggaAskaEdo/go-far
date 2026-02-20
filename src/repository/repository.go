package repository

import (
	"time"

	"go-far/src/config"
	"go-far/src/repository/car"
	"go-far/src/repository/user"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	User user.UserRepositoryItf
	Car  car.CarRepositoryItf
}

func InitRepository(sql0 *sqlx.DB, redis0 *redis.Client, queryLoader *config.QueryLoader, cacheTTL time.Duration) *Repository {
	return &Repository{
		User: user.InitUserRepository(
			sql0,
			redis0,
			queryLoader,
			cacheTTL,
		),
		Car: car.InitCarRepository(
			sql0,
			redis0,
			queryLoader,
			cacheTTL,
		),
	}
}
