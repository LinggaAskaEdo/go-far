package car

import (
	"context"
	"errors"
	"time"

	"go-far/internal/model/entity"
	appErr "go-far/internal/model/errors"
	"go-far/internal/util"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

func (r *carRepository) createSQLCar(ctx context.Context, tx pgx.Tx, car *entity.Car) error {
	query, args, err := r.queryLoader.Compile("CreateCar", car)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_create_car_query_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_create_car_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	err = tx.QueryRow(ctx, query, args...).Scan(&car.ID, &car.Brand, &car.Model, &car.Year, &car.Color, &car.LicensePlate, &car.IsAvailable, &car.CreatedAt, &car.UpdatedAt)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("create_car_err")
		return appErr.Wrap(err, "create_car_err")
	}

	return nil
}

func (r *carRepository) createBulkSQLCars(ctx context.Context, tx pgx.Tx, cars []*entity.Car) error {
	if len(cars) == 0 {
		return appErr.NewWithCode(appErr.CodeHTTPBadRequest, "no cars to create")
	}

	now := time.Now()
	for _, car := range cars {
		car.CreatedAt = now
		car.UpdatedAt = now
	}

	query, args, err := r.queryLoader.Compile("CreateCarBulk", cars)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_create_bulk_cars_query_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_create_bulk_cars_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("create_bulk_cars_err")
		return appErr.Wrap(err, "create_bulk_cars_err")
	}

	return nil
}

func (r *carRepository) updateSQLCar(ctx context.Context, id uuid.UUID, car *entity.Car) error {
	data := map[string]any{
		"ID":           id,
		"Brand":        car.Brand,
		"Model":        car.Model,
		"Year":         car.Year,
		"Color":        car.Color,
		"LicensePlate": car.LicensePlate,
		"IsAvailable":  car.IsAvailable,
		"UpdatedAt":    time.Now(),
	}

	query, args, err := r.queryLoader.Compile("UpdateCar", data)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_update_car_query_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_update_car_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	var updatedAt time.Time
	err = r.sql0.QueryRow(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows { //nolint:errorlint
			zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_not_found_for_update")
			return appErr.NewWithCode(appErr.CodeSQLEmptyRow, "car_not_found_for_update")
		}
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id.String()).Msg("update_car_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLUpdate, "update_car_err")
	}

	return nil
}

func (r *carRepository) deleteSQLCar(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.queryLoader.Compile("DeleteCar", map[string]any{"ID": id})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_delete_car_query_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_delete_car_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	result, err := r.sql0.Exec(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id.String()).Msg("delete_car_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLDelete, "delete_car_err")
	}

	rows := result.RowsAffected()

	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_not_found_for_deletion")
		return appErr.NewWithCode(appErr.CodeSQLEmptyRow, "car_not_found_for_deletion")
	}

	return nil
}

func (r *carRepository) findCarSQLByID(ctx context.Context, id uuid.UUID) (*entity.Car, error) {
	query, args, err := r.queryLoader.Compile("FindCarByID", map[string]any{"ID": id})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_car_query_err")
		return nil, appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_find_car_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	var car entity.Car
	err = r.sql0.QueryRow(ctx, query, args...).Scan(&car)
	if err != nil {
		if err == pgx.ErrNoRows { //nolint:errorlint
			return nil, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "find_car_err")
		}

		zerolog.Ctx(ctx).Error().Err(err).Str("id", id.String()).Msg("find_car_err")
		return nil, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "find_car_err")
	}

	return &car, nil
}

func (r *carRepository) findCarByUserIDSQL(ctx context.Context, userID uuid.UUID) ([]*entity.Car, error) {
	query, args, err := r.queryLoader.Compile("FindCarsByUserID", map[string]any{"UserID": userID})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_cars_query_err")
		return nil, appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_find_cars_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	rows, err := r.sql0.Query(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Msg("find_cars_by_user_err")
		return nil, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "find_cars_by_user_err")
	}
	defer rows.Close()

	var cars []*entity.Car
	for rows.Next() {
		var car entity.Car
		if err := rows.Scan(&car.ID, &car.Brand, &car.Model, &car.Year, &car.Color, &car.LicensePlate, &car.IsAvailable, &car.CreatedAt, &car.UpdatedAt); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("scan_car_err")
			return nil, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "scan_car_err")
		}
		cars = append(cars, &car)
	}

	return cars, nil
}

func (r *carRepository) countCarsByUserIDSQL(ctx context.Context, userID uuid.UUID) (int, error) {
	query, args, err := r.queryLoader.Compile("CountCarsByUserID", map[string]any{"UserID": userID})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_count_cars_query_err")
		return 0, appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_count_cars_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	var count int
	err = r.sql0.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Msg("count_cars_by_user_err")
		return 0, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "count_cars_by_user_err")
	}

	return count, nil
}

func (r *carRepository) assignCarToUserSQL(ctx context.Context, userID, carID uuid.UUID) error {
	query, args, err := r.queryLoader.Compile("AssignCarToUser", map[string]any{
		"UserID": userID,
		"CarID":  carID,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_assign_car_to_user_query_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_assign_car_to_user_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	_, err = r.sql0.Exec(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Str("car_id", carID.String()).Msg("assign_car_to_user_err")
		return appErr.Wrap(err, "assign_car_to_user_err")
	}

	return nil
}

func (r *carRepository) assignCarsToUserBulkSQL(ctx context.Context, userID uuid.UUID, carIDs []uuid.UUID) error {
	userCars := make([]entity.UserCar, len(carIDs))
	for i, carID := range carIDs {
		userCars[i] = entity.UserCar{UserID: userID, CarID: carID}
	}

	query, args, err := r.queryLoader.Compile("AssignCarToUserBulk", userCars)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_assign_cars_bulk_query_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_assign_cars_bulk_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	_, err = r.sql0.Exec(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", userID.String()).Msg("assign_cars_to_user_bulk_err")
		return appErr.Wrap(err, "assign_cars_to_user_bulk_err")
	}

	return nil
}

func (r *carRepository) findCarByIDWithOwnerSQL(ctx context.Context, id uuid.UUID) (*entity.CarWithOwner, error) {
	query, args, err := r.queryLoader.Compile("FindCarByIDWithOwner", map[string]any{"ID": id})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_car_with_owner_query_err")
		return nil, appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_find_car_with_owner_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	var carWithOwner entity.CarWithOwner
	err = r.sql0.QueryRow(ctx, query, args...).Scan(&carWithOwner)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_not_found")
			return nil, appErr.WrapWithCode(err, appErr.CodeSQLEmptyRow, "car_not_found")
		}

		zerolog.Ctx(ctx).Error().Err(err).Str("id", id.String()).Msg("find_car_with_owner_err")
		return nil, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "find_car_with_owner_err")
	}

	return &carWithOwner, nil
}

func (r *carRepository) transferOwnershipSQL(ctx context.Context, carID, newUserID uuid.UUID) error {
	data := map[string]any{
		"CarID":     carID,
		"NewUserID": newUserID,
	}

	query, args, err := r.queryLoader.Compile("TransferCarOwnership", data)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_transfer_ownership_query_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_transfer_ownership_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	result, err := r.sql0.Exec(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("car_id", carID.String()).Msg("transfer_ownership_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLUpdate, "transfer_ownership_err")
	}

	rows := result.RowsAffected()

	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("car_id", carID.String()).Msg("car_not_found_for_transfer")
		return appErr.NewWithCode(appErr.CodeSQLEmptyRow, "car_not_found_for_transfer")
	}

	return nil
}

func (r *carRepository) bulkUpdateAvailabilitySQL(ctx context.Context, carIDs []uuid.UUID, isAvailable bool) error {
	if len(carIDs) == 0 {
		return nil
	}

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
		return appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_bulk_update_availability_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	result, err := r.sql0.Exec(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("bulk_update_availability_err")
		return appErr.WrapWithCode(err, appErr.CodeSQLUpdate, "bulk_update_availability_err")
	}

	rows := result.RowsAffected()

	zerolog.Ctx(ctx).Debug().Int64("rows_affected", rows).Msg("bulk_update_availability_success")

	return nil
}

func (r *carRepository) checkCarOwnershipSQL(ctx context.Context, carID uuid.UUID, userID string) (bool, error) {
	query, args, err := r.queryLoader.Compile("CheckCarOwnership", map[string]any{
		"CarID":  carID,
		"UserID": userID,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_check_car_ownership_query_err")
		return false, appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_check_car_ownership_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	var count int
	err = r.sql0.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		zerolog.Ctx(ctx).Error().Err(err).Str("car_id", carID.String()).Str("user_id", userID).Msg("check_car_ownership_err")
		return false, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "check_car_ownership_err")
	}

	return count > 0, nil
}

func (r *carRepository) checkCarsOwnershipSQL(ctx context.Context, carIDs []uuid.UUID, userID string) (map[uuid.UUID]bool, error) {
	if len(carIDs) == 0 {
		return make(map[uuid.UUID]bool), nil
	}

	query, args, err := r.queryLoader.Compile("CheckCarsOwnership", map[string]any{
		"CarIDs": carIDs,
		"UserID": userID,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_check_cars_ownership_query_err")
		return nil, appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_check_cars_ownership_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(query)).Any("args", args).Msg("compiled_query")

	rows, err := r.sql0.Query(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("check_cars_ownership_err")
		return nil, appErr.WrapWithCode(err, appErr.CodeSQLRead, "check_cars_ownership_err")
	}
	defer rows.Close()

	ownedSet := make(map[uuid.UUID]struct{})
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("scan_owned_car_err")
			return nil, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "scan_owned_car_err")
		}
		ownedSet[id] = struct{}{}
	}

	ownershipMap := make(map[uuid.UUID]bool, len(carIDs))
	for _, id := range carIDs {
		_, ownershipMap[id] = ownedSet[id]
	}

	return ownershipMap, nil
}
