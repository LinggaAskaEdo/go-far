package car

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go-far/src/model/entity"
	x "go-far/src/model/errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func (r *carRepository) Create(ctx context.Context, car *entity.Car) error {
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

func (r *carRepository) CreateBulk(ctx context.Context, cars []*entity.Car) error {
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

func (r *carRepository) AssignCarToUser(ctx context.Context, userID uuid.UUID, carID uuid.UUID) error {
	query, args, err := r.queryLoader.Compile("AssignCarToUser", map[string]any{
		"UserID": userID,
		"CarID":  carID,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_assign_car_to_user_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_assign_car_to_user_query_err")
	}

	_, err = r.sql0.ExecContext(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Str("car_id", carID.String()).Msg("assign_car_to_user_err")
		return x.Wrap(err, "assign_car_to_user_err")
	}

	return nil
}

func (r *carRepository) AssignCarsToUserBulk(ctx context.Context, userID uuid.UUID, carIDs []uuid.UUID) error {
	type userCar struct {
		UserID uuid.UUID
		CarID  uuid.UUID
	}

	userCars := make([]userCar, len(carIDs))
	for i, carID := range carIDs {
		userCars[i] = userCar{UserID: userID, CarID: carID}
	}

	query, args, err := r.queryLoader.Compile("AssignCarToUserBulk", userCars)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_assign_cars_bulk_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_assign_cars_bulk_query_err")
	}

	_, err = r.sql0.ExecContext(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Msg("assign_cars_to_user_bulk_err")
		return x.Wrap(err, "assign_cars_to_user_bulk_err")
	}

	return nil
}

func (r *carRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Car, error) {
	var car entity.Car

	cacheKey := fmt.Sprintf("car:%s", id.String())

	cached, err := r.redis0.Get(ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cached), &car); err == nil {
			zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_found_in_cache")
			return &car, nil
		}
	}

	query, args, err := r.queryLoader.Compile("FindCarByID", map[string]any{"ID": id})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_car_query_err")
		return nil, x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_find_car_query_err")
	}

	err = r.sql0.GetContext(ctx, &car, query, args...)
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

func (r *carRepository) FindByIDWithOwner(ctx context.Context, id uuid.UUID) (*entity.CarWithOwner, error) {
	var carWithOwner entity.CarWithOwner

	query, args, err := r.queryLoader.Compile("FindCarByIDWithOwner", map[string]any{"ID": id})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_car_with_owner_query_err")
		return nil, x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_find_car_with_owner_query_err")
	}

	err = r.sql0.GetContext(ctx, &carWithOwner, query, args...)
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

func (r *carRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Car, error) {
	var cars []*entity.Car

	query, args, err := r.queryLoader.Compile("FindCarsByUserID", map[string]any{"UserID": userID})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_cars_query_err")
		return nil, x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_find_cars_query_err")
	}

	err = r.sql0.SelectContext(ctx, &cars, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Msg("find_cars_by_user_err")
		return nil, x.WrapWithCode(err, x.CodeSQLRowScan, "find_cars_by_user_err")
	}

	return cars, nil
}

func (r *carRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int

	query, args, err := r.queryLoader.Compile("CountCarsByUserID", map[string]any{"UserID": userID})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_count_cars_query_err")
		return 0, x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_count_cars_query_err")
	}

	err = r.sql0.GetContext(ctx, &count, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Msg("count_cars_by_user_err")
		return 0, x.WrapWithCode(err, x.CodeSQLRowScan, "count_cars_by_user_err")
	}

	return count, nil
}

func (r *carRepository) Update(ctx context.Context, id uuid.UUID, car *entity.Car) error {
	return r.updateSQLCar(ctx, id, car)
}

func (r *carRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.deleteSQLCar(ctx, id)
}

func (r *carRepository) TransferOwnership(ctx context.Context, carID, newUserID uuid.UUID) error {
	data := map[string]any{
		"CarID":     carID,
		"NewUserID": newUserID,
	}

	query, args, err := r.queryLoader.Compile("TransferCarOwnership", data)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_transfer_ownership_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_transfer_ownership_query_err")
	}

	result, err := r.sql0.ExecContext(ctx, query, args...)
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
	// Convert []uuid.UUID to []string for template
	idStrs := make([]string, len(carIDs))
	for i, id := range carIDs {
		idStrs[i] = id.String()
	}

	data := map[string]any{
		"CarIDs":      idStrs,
		"IsAvailable": isAvailable,
		"UpdatedAt":   time.Now(),
	}

	query, args, err := r.queryLoader.Compile("BulkUpdateCarAvailability", data)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_bulk_update_availability_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_bulk_update_availability_query_err")
	}

	result, err := r.sql0.ExecContext(ctx, query, args...)
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

	// Invalidate cache for affected cars
	for _, id := range carIDs {
		cacheKey := "car:" + id.String()
		r.redis0.Del(ctx, cacheKey)
	}

	return nil
}
