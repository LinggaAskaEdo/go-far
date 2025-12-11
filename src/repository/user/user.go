package user

import (
	"context"
	"time"

	"go-far/src/config"
	"go-far/src/domain"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type UserRepositoryItf interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindAll(ctx context.Context, filter domain.UserFilter) ([]*domain.User, int64, error)
	Update(ctx context.Context, id string, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db          *sqlx.DB
	redis       *redis.Client
	queryLoader *config.QueryLoader
	cacheTTL    time.Duration
}

func InitUserRepository(db *sqlx.DB, redis *redis.Client, queryLoader *config.QueryLoader, cacheTTL time.Duration) UserRepositoryItf {
	return &userRepository{
		db:          db,
		redis:       redis,
		queryLoader: queryLoader,
		cacheTTL:    cacheTTL,
	}
}
