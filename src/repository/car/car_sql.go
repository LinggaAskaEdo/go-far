package car

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-far/src/domain"
	x "go-far/src/errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

func (r *carRepository) createSQLCar(ctx context.Context, tx *sqlx.Tx, car *domain.Car) error {
	query, _ := r.queryLoader.Get("CreateCar")

	err := tx.QueryRowContext(
		ctx,
		query,
		car.UserID,
		car.Brand,
		car.Model,
		car.Year,
		car.Color,
		car.LicensePlate,
		car.IsAvailable,
	).Scan(&car.ID, &car.CreatedAt, &car.UpdatedAt)

	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("user_id", car.UserID).Msg("create_car_err")
		return x.Wrap(err, "create_car_err")
	}

	return nil
}

func (r *carRepository) createBulkSQLCars(ctx context.Context, tx *sqlx.Tx, cars []*domain.Car) error {
	if len(cars) == 0 {
		return x.NewWithCode(x.CodeHTTPBadRequest, "no cars to create")
	}

	now := time.Now()
	templateData := make([]map[string]any, len(cars))

	for i, car := range cars {
		car.CreatedAt = now
		car.UpdatedAt = now
		templateData[i] = map[string]any{
			"user_id":       car.UserID,
			"brand":         car.Brand,
			"model":         car.Model,
			"year":          car.Year,
			"color":         car.Color,
			"license_plate": car.LicensePlate,
			"is_available":  car.IsAvailable,
			"created_at":    car.CreatedAt,
			"updated_at":    car.UpdatedAt,
		}
	}

	query, args, err := r.queryLoader.ExecuteTemplate("CreateCarBulk", templateData)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_create_bulk_cars_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_create_bulk_cars_query_err")
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("create_bulk_cars_err")
		return x.Wrap(err, "create_bulk_cars_err")
	}

	return nil
}

func (r *carRepository) updateSQLCar(ctx context.Context, id uuid.UUID, car *domain.Car) error {
	query, _ := r.queryLoader.Get("UpdateCar")

	var updatedAt time.Time
	err := r.sql0.QueryRowContext(
		ctx,
		query,
		car.Brand,
		car.Model,
		car.Year,
		car.Color,
		car.LicensePlate,
		car.IsAvailable,
		time.Now(),
		id.String(),
	).Scan(&updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_not_found_for_update")
			return x.NewWithCode(x.CodeSQLEmptyRow, "car_not_found_for_update")
		}
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id.String()).Msg("update_car_err")
		return x.WrapWithCode(err, x.CodeSQLUpdate, "update_car_err")
	}

	cacheKey := fmt.Sprintf("car:%s", id.String())
	r.redis0.Del(ctx, cacheKey)

	return nil
}

func (r *carRepository) deleteSQLCar(ctx context.Context, id uuid.UUID) error {
	query, _ := r.queryLoader.Get("DeleteCar")

	result, err := r.sql0.ExecContext(ctx, query, id.String())
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id.String()).Msg("delete_car_err")
		return x.WrapWithCode(err, x.CodeSQLDelete, "delete_car_err")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id.String()).Msg("failed_to_get_rows_affected")
		return x.WrapWithCode(err, x.CodeSQLDelete, "failed_to_get_rows_affected")
	}

	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("id", id.String()).Msg("car_not_found_for_deletion")
		return x.NewWithCode(x.CodeSQLEmptyRow, "car_not_found_for_deletion")
	}

	cacheKey := fmt.Sprintf("car:%s", id.String())
	r.redis0.Del(ctx, cacheKey)

	return nil
}
