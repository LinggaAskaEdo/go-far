package car

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-far/src/model/entity"
	x "go-far/src/model/errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

func (r *carRepository) createSQLCar(ctx context.Context, tx *sqlx.Tx, car *entity.Car) error {
	query, args, err := r.queryLoader.Compile("CreateCar", car)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_create_car_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_create_car_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", query).Any("args", args).Msg("compiled_query")

	err = tx.QueryRowContext(ctx, query, args...).Scan(&car.ID, &car.CreatedAt, &car.UpdatedAt)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("create_car_err")
		return x.Wrap(err, "create_car_err")
	}

	return nil
}

func (r *carRepository) createBulkSQLCars(ctx context.Context, tx *sqlx.Tx, cars []*entity.Car) error {
	if len(cars) == 0 {
		return x.NewWithCode(x.CodeHTTPBadRequest, "no cars to create")
	}

	now := time.Now()
	for _, car := range cars {
		car.CreatedAt = now
		car.UpdatedAt = now
	}

	query, args, err := r.queryLoader.Compile("CreateCarBulk", cars)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_create_bulk_cars_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_create_bulk_cars_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", query).Any("args", args).Msg("compiled_query")

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("create_bulk_cars_err")
		return x.Wrap(err, "create_bulk_cars_err")
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
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_update_car_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", query).Any("args", args).Msg("compiled_query")

	var updatedAt time.Time
	err = r.sql0.QueryRowContext(ctx, query, args...).Scan(&updatedAt)
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
	query, args, err := r.queryLoader.Compile("DeleteCar", map[string]any{"ID": id})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_delete_car_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_delete_car_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", query).Any("args", args).Msg("compiled_query")

	result, err := r.sql0.ExecContext(ctx, query, args...)
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
