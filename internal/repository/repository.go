package repository

import (
	"time"

	"go-far/internal/infra/query"
	"go-far/internal/repository/car"
	"go-far/internal/repository/user"

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
