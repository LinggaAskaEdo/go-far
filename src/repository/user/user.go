package user

import (
	"context"
	"time"

	"go-far/src/config"
	"go-far/src/domain"
	"go-far/src/dto"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type UserRepositoryItf interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	FindUserByID(ctx context.Context, id string) (domain.User, error)
	FindAllUser(ctx context.Context, cacheControl dto.CacheControl, filter dto.UserFilter) ([]domain.User, dto.Pagination, error)
	UpdateUser(ctx context.Context, id string, user domain.User) error
	DeleteUser(ctx context.Context, id string) error
}

type userRepository struct {
	sql0        *sqlx.DB
	redis0      *redis.Client
	queryLoader *config.QueryLoader
	cacheTTL    time.Duration
}

func InitUserRepository(sql0 *sqlx.DB, redis0 *redis.Client, queryLoader *config.QueryLoader, cacheTTL time.Duration) UserRepositoryItf {
	return &userRepository{
		sql0:        sql0,
		redis0:      redis0,
		queryLoader: queryLoader,
		cacheTTL:    cacheTTL,
	}
}
