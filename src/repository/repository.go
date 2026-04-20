package repository

import (
	"time"

	"go-far/src/config/query"
	"go-far/src/repository/car"
	"go-far/src/repository/user"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	User user.UserRepositoryItf
	Car  car.CarRepositoryItf
}

func InitRepository(sql0 *pgxpool.Pool, redis0 *redis.Client, queryLoader *query.QueryLoader, cacheTTL time.Duration) *Repository {
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
