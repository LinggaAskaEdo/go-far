package repository

import (
	"time"

	"go-far/src/config"
	"go-far/src/repository/user"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	User user.UserRepositoryItf
}

func InitRepository(db *sqlx.DB, redis *redis.Client, queryLoader *config.QueryLoader, cacheTTL time.Duration) *Repository {
	return &Repository{
		User: user.InitUserRepository(
			db,
			redis,
			queryLoader,
			cacheTTL,
		),
	}
}
