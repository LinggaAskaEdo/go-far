package car

import (
	"context"
	"time"

	"go-far/internal/infra/query"
	"go-far/internal/model/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type CarRepositoryItf interface {
	Create(ctx context.Context, car *entity.Car) error
	CreateBulk(ctx context.Context, cars []*entity.Car) error
	AssignCarToUser(ctx context.Context, userID uuid.UUID, carID uuid.UUID) error
	AssignCarsToUserBulk(ctx context.Context, userID uuid.UUID, carIDs []uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Car, error)
	FindByIDWithOwner(ctx context.Context, id uuid.UUID) (*entity.CarWithOwner, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Car, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	Update(ctx context.Context, id uuid.UUID, car *entity.Car) error
	Delete(ctx context.Context, id uuid.UUID) error
	TransferOwnership(ctx context.Context, carID, newUserID uuid.UUID) error
	BulkUpdateAvailability(ctx context.Context, carIDs []uuid.UUID, isAvailable bool) error
	IsCarOwnedByUser(ctx context.Context, carID uuid.UUID, userID string) (bool, error)
	AreCarsOwnedByUser(ctx context.Context, carIDs []uuid.UUID, userID string) (map[uuid.UUID]bool, error)
}

type carRepository struct {
	sql0        *pgxpool.Pool
	redis0      *redis.Client
	queryLoader *query.QueryLoader
	cacheTTL    time.Duration
}

func InitCarRepository(sql0 *pgxpool.Pool, redis0 *redis.Client, queryLoader *query.QueryLoader, cacheTTL time.Duration) CarRepositoryItf {
	return &carRepository{
		sql0:        sql0,
		redis0:      redis0,
		queryLoader: queryLoader,
		cacheTTL:    cacheTTL,
	}
}
