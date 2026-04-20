package car

import (
	"context"
	"encoding/json"
	"fmt"

	"go-far/src/model/entity"
	x "go-far/src/model/errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const cacheKeyCar = "car:%s"

func (r *carRepository) Create(ctx context.Context, car *entity.Car) error {
	tx, err := r.sql0.Begin(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("tx_create_car")
		return x.Wrap(err, "tx_create_car")
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				zerolog.Ctx(ctx).Error().Err(rollbackErr).Msg("rollback_create_car")
			}
		}
	}()

	if err = r.createSQLCar(ctx, tx, car); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("sql_create_car")
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("commit_create_car")
		return x.Wrap(err, "commit_create_car")
	}

	return nil
}

func (r *carRepository) CreateBulk(ctx context.Context, cars []*entity.Car) error {
	tx, err := r.sql0.Begin(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("tx_create_bulk_cars")
		return x.Wrap(err, "tx_create_bulk_cars")
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				zerolog.Ctx(ctx).Error().Err(rollbackErr).Msg("rollback_create_bulk_cars")
			}
		}
	}()

	if err = r.createBulkSQLCars(ctx, tx, cars); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("sql_create_bulk_cars")
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("commit_create_bulk_cars")
		return x.Wrap(err, "commit_create_bulk_cars")
	}

	return nil
}

func (r *carRepository) AssignCarToUser(ctx context.Context, userID uuid.UUID, carID uuid.UUID) error {
	return r.assignCarToUserSQL(ctx, userID, carID)
}

func (r *carRepository) AssignCarsToUserBulk(ctx context.Context, userID uuid.UUID, carIDs []uuid.UUID) error {
	return r.assignCarsToUserBulkSQL(ctx, userID, carIDs)
}

func (r *carRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Car, error) {
	cacheKey := fmt.Sprintf(cacheKeyCar, id.String())

	cached, err := r.redis0.Get(ctx, cacheKey).Result()
	if err == nil {
		var car entity.Car
		if err := json.Unmarshal([]byte(cached), &car); err == nil {
			zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_found_in_cache")
			return &car, nil
		}
	}

	car, err := r.findCarSQLByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(car)
	r.redis0.Set(ctx, cacheKey, data, r.cacheTTL)

	return car, nil
}

func (r *carRepository) FindByIDWithOwner(ctx context.Context, id uuid.UUID) (*entity.CarWithOwner, error) {
	return r.findCarByIDWithOwnerSQL(ctx, id)
}

func (r *carRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Car, error) {
	return r.findCarByUserIDSQL(ctx, userID)
}

func (r *carRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.countCarsByUserIDSQL(ctx, userID)
}

func (r *carRepository) Update(ctx context.Context, id uuid.UUID, car *entity.Car) error {
	err := r.updateSQLCar(ctx, id, car)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf(cacheKeyCar, id.String())
	r.redis0.Del(ctx, cacheKey)

	return nil
}

func (r *carRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.deleteSQLCar(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf(cacheKeyCar, id.String())
	r.redis0.Del(ctx, cacheKey)

	return nil
}

func (r *carRepository) TransferOwnership(ctx context.Context, carID, newUserID uuid.UUID) error {
	err := r.transferOwnershipSQL(ctx, carID, newUserID)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf(cacheKeyCar, carID.String())
	r.redis0.Del(ctx, cacheKey)

	return nil
}

func (r *carRepository) BulkUpdateAvailability(ctx context.Context, carIDs []uuid.UUID, isAvailable bool) error {
	err := r.bulkUpdateAvailabilitySQL(ctx, carIDs, isAvailable)
	if err != nil {
		return err
	}

	for _, id := range carIDs {
		cacheKey := fmt.Sprintf(cacheKeyCar, id.String())
		r.redis0.Del(ctx, cacheKey)
	}

	return nil
}

func (r *carRepository) IsCarOwnedByUser(ctx context.Context, carID uuid.UUID, userID string) (bool, error) {
	return r.checkCarOwnershipSQL(ctx, carID, userID)
}

func (r *carRepository) AreCarsOwnedByUser(ctx context.Context, carIDs []uuid.UUID, userID string) (map[uuid.UUID]bool, error) {
	ownershipMap, err := r.checkCarsOwnershipSQL(ctx, carIDs, userID)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("check_cars_ownership_err")
		return nil, err
	}

	return ownershipMap, nil
}
