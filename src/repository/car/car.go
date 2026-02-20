package car

import (
	"go-far/src/config"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type CarRepositoryItf interface {
	// Create(ctx context.Context, car *domain.Car) error
	// CreateBulk(ctx context.Context, cars []*domain.Car) error
	// FindByID(ctx context.Context, id uuid.UUID) (*domain.Car, error)
	// FindByIDWithOwner(ctx context.Context, id uuid.UUID) (*domain.CarWithOwner, error)
	// FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Car, error)
	// CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	// Update(ctx context.Context, id uuid.UUID, car *domain.Car) error
	// Delete(ctx context.Context, id uuid.UUID) error
	// TransferOwnership(ctx context.Context, carID, newUserID uuid.UUID) error
	// BulkUpdateAvailability(ctx context.Context, carIDs []uuid.UUID, isAvailable bool) error
	// WithTx(tx *sql.Tx) CarRepository
}

type carRepository struct {
	sql0        *sqlx.DB
	redis0      *redis.Client
	queryLoader *config.QueryLoader
	cacheTTL    time.Duration
}

func InitCarRepository(sql0 *sqlx.DB, redis0 *redis.Client, queryLoader *config.QueryLoader, cacheTTL time.Duration) CarRepositoryItf {
	return &carRepository{
		sql0:        sql0,
		redis0:      redis0,
		queryLoader: queryLoader,
		cacheTTL:    cacheTTL,
	}
}
