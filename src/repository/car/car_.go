package car

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go-far/src/domain"
	x "go-far/src/errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func (r *carRepository) Create(ctx context.Context, car *domain.Car) error {
	tx, err := r.sql0.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("tx_create_car")
		return x.Wrap(err, "tx_create_car")
	}

	if err = r.createSQLCar(ctx, tx, car); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			zerolog.Ctx(ctx).Error().Err(rollbackErr).Msg("rollback_create_car")
		}
		zerolog.Ctx(ctx).Error().Err(err).Msg("sql_create_car")
		return err
	}

	if err = tx.Commit(); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			zerolog.Ctx(ctx).Error().Err(rollbackErr).Msg("rollback_after_commit_failure")
		}
		zerolog.Ctx(ctx).Error().Err(err).Msg("commit_create_car")
		return x.Wrap(err, "commit_create_car")
	}

	return nil
}

func (r *carRepository) CreateBulk(ctx context.Context, cars []*domain.Car) error {
	tx, err := r.sql0.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("tx_create_bulk_cars")
		return x.Wrap(err, "tx_create_bulk_cars")
	}

	if err = r.createBulkSQLCars(ctx, tx, cars); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			zerolog.Ctx(ctx).Error().Err(rollbackErr).Msg("rollback_create_bulk_cars")
		}
		zerolog.Ctx(ctx).Error().Err(err).Msg("sql_create_bulk_cars")
		return err
	}

	if err = tx.Commit(); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			zerolog.Ctx(ctx).Error().Err(rollbackErr).Msg("rollback_after_commit_failure")
		}
		zerolog.Ctx(ctx).Error().Err(err).Msg("commit_create_bulk_cars")
		return x.Wrap(err, "commit_create_bulk_cars")
	}

	return nil
}

func (r *carRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Car, error) {
	var car domain.Car

	cacheKey := fmt.Sprintf("car:%s", id.String())

	cached, err := r.redis0.Get(ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cached), &car); err == nil {
			zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_found_in_cache")
			return &car, nil
		}
	}

	query, _ := r.queryLoader.Get("FindCarByID")

	err = r.sql0.GetContext(ctx, &car, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_not_found")
			return nil, x.WrapWithCode(err, x.CodeSQLEmptyRow, "car_not_found")
		}
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id.String()).Msg("find_car_err")
		return nil, x.WrapWithCode(err, x.CodeSQLRowScan, "find_car_err")
	}

	data, _ := json.Marshal(car)
	r.redis0.Set(ctx, cacheKey, data, r.cacheTTL)

	return &car, nil
}

func (r *carRepository) FindByIDWithOwner(ctx context.Context, id uuid.UUID) (*domain.CarWithOwner, error) {
	var carWithOwner domain.CarWithOwner

	query, _ := r.queryLoader.Get("FindCarByIDWithOwner")

	err := r.sql0.GetContext(ctx, &carWithOwner, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_not_found")
			return nil, x.WrapWithCode(err, x.CodeSQLEmptyRow, "car_not_found")
		}
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id.String()).Msg("find_car_with_owner_err")
		return nil, x.WrapWithCode(err, x.CodeSQLRowScan, "find_car_with_owner_err")
	}

	return &carWithOwner, nil
}

func (r *carRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Car, error) {
	var cars []*domain.Car

	query, _ := r.queryLoader.Get("FindCarsByUserID")

	err := r.sql0.SelectContext(ctx, &cars, query, userID.String())
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Msg("find_cars_by_user_err")
		return nil, x.WrapWithCode(err, x.CodeSQLRowScan, "find_cars_by_user_err")
	}

	return cars, nil
}

func (r *carRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int

	query, _ := r.queryLoader.Get("CountCarsByUserID")

	err := r.sql0.GetContext(ctx, &count, query, userID.String())
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Msg("count_cars_by_user_err")
		return 0, x.WrapWithCode(err, x.CodeSQLRowScan, "count_cars_by_user_err")
	}

	return count, nil
}

func (r *carRepository) Update(ctx context.Context, id uuid.UUID, car *domain.Car) error {
	err := r.updateSQLCar(ctx, id, car)
	if err != nil {
		return err
	}

	return nil
}

func (r *carRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.deleteSQLCar(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *carRepository) TransferOwnership(ctx context.Context, carID, newUserID uuid.UUID) error {
	query, _ := r.queryLoader.Get("TransferCarOwnership")

	result, err := r.sql0.ExecContext(
		ctx,
		query,
		newUserID.String(),
		time.Now(),
		carID.String(),
	)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("car_id", carID.String()).Msg("transfer_ownership_err")
		return x.WrapWithCode(err, x.CodeSQLUpdate, "transfer_ownership_err")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("car_id", carID.String()).Msg("failed_to_get_rows_affected")
		return x.WrapWithCode(err, x.CodeSQLUpdate, "failed_to_get_rows_affected")
	}

	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("car_id", carID.String()).Msg("car_not_found_for_transfer")
		return x.NewWithCode(x.CodeSQLEmptyRow, "car_not_found_for_transfer")
	}

	cacheKey := fmt.Sprintf("car:%s", carID.String())
	r.redis0.Del(ctx, cacheKey)

	return nil
}

func (r *carRepository) BulkUpdateAvailability(ctx context.Context, carIDs []uuid.UUID, isAvailable bool) error {
	query, _ := r.queryLoader.Get("BulkUpdateCarAvailability")

	result, err := r.sql0.ExecContext(
		ctx,
		query,
		isAvailable,
		time.Now(),
		carIDs,
	)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("bulk_update_availability_err")
		return x.WrapWithCode(err, x.CodeSQLUpdate, "bulk_update_availability_err")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed_to_get_rows_affected")
		return x.WrapWithCode(err, x.CodeSQLUpdate, "failed_to_get_rows_affected")
	}

	zerolog.Ctx(ctx).Debug().Int64("rows_affected", rows).Msg("bulk_update_availability_success")

	return nil
}
