package user

import (
	"context"
	"time"

	"go-far/src/config/query"
	"go-far/src/model/dto"
	"go-far/src/model/entity"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type UserRepositoryItf interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindAll(ctx context.Context, cacheControl dto.CacheControl, filter dto.UserFilter) (*[]entity.User, *dto.Pagination, error)
	FindAllV2(ctx context.Context, filter dto.UserFilterV2) (*[]entity.User, *dto.Pagination, error)
	Update(ctx context.Context, id string, user *entity.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	sql0        *sqlx.DB
	redis0      *redis.Client
	queryLoader *query.QueryLoader
	cacheTTL    time.Duration
}

func InitUserRepository(sql0 *sqlx.DB, redis0 *redis.Client, queryLoader *query.QueryLoader, cacheTTL time.Duration) UserRepositoryItf {
	return &userRepository{
		sql0:        sql0,
		redis0:      redis0,
		queryLoader: queryLoader,
		cacheTTL:    cacheTTL,
	}
}
