package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go-far/src/domain"
	"go-far/src/dto"
	x "go-far/src/errors"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func (d *userRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	tx, err := d.sql0.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("tx_create_user")
		return user, x.Wrap(err, "tx_create_user")
	}

	tx, user, err = d.createSQLUser(ctx, tx, user)
	if err != nil {
		_ = tx.Rollback()
		zerolog.Ctx(ctx).Error().Err(err).Msg("sql_create_user")
		return user, x.Wrap(err, "sql_create_user")
	}

	if err = tx.Commit(); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("commit_create_user")
		return user, x.Wrap(err, "commit_create_user")
	}

	return user, nil
}

func (d *userRepository) FindUserByID(ctx context.Context, id string) (domain.User, error) {
	var user domain.User

	cacheKey := fmt.Sprintf("user:%s", id)

	cached, err := d.redis0.Get(ctx, cacheKey).Result()
	if err == nil {

		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			zerolog.Ctx(ctx).Debug().Str("id", id).Msg("data_found_in_cache")
			return user, nil
		}
	}

	query, _ := d.queryLoader.Get("FindUserByID")

	err = d.sql0.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			zerolog.Ctx(ctx).Debug().Str("id", id).Msg("user_not_found")
			return user, x.WrapWithCode(err, x.CodeSQLEmptyRow, "user_not_found")
		}

		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("find_user_err")
		return user, x.WrapWithCode(err, x.CodeSQLRowScan, "find_user_err")
	}

	data, _ := json.Marshal(user)
	d.redis0.Set(ctx, cacheKey, data, d.cacheTTL)

	return user, nil
}

func (d *userRepository) FindAllUser(ctx context.Context, cacheControl dto.CacheControl, filter dto.UserFilter) ([]domain.User, dto.Pagination, error) {
	if cacheControl.MustRevalidate {
		result, pagination, err := d.findAllSQLUser(ctx, filter)
		if err != nil {
			return result, pagination, err
		}

		if err = d.setCacheFindAllUser(ctx, filter, result, pagination); err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Send()
		}

		return result, pagination, nil
	}

	result, pagination, err := d.getCacheFindAllUser(ctx, filter)
	if err == redis.Nil {
		zerolog.Ctx(ctx).Warn().Err(err).Send()

		result, pagination, err = d.findAllSQLUser(ctx, filter)
		if err != nil {
			return result, pagination, err
		}

		if err = d.setCacheFindAllUser(ctx, filter, result, pagination); err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Send()
		}

		return result, pagination, nil
	} else if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Send()

		// fallback if there is redis error e.g. bad conn, etc.
		// this is quite critical during high load traffic since it could be
		// thundering our db. (thundering herd).
		// we leave as it is to reduce code complexity [TODO LATER]
		return d.findAllSQLUser(ctx, filter)
	}

	return result, pagination, nil
}

func (d *userRepository) UpdateUser(ctx context.Context, id string, user domain.User) error {
	query, _ := d.queryLoader.Get("UpdateUser")

	result, err := d.sql0.ExecContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Age,
		time.Now(),
		id,
	)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("update_user_err")
		return x.WrapWithCode(err, x.CodeSQLUpdate, "update_user_err")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("id", id).Msg("user_not_found_for_update")
		return x.WrapWithCode(err, x.CodeSQLEmptyRow, "user_not_found_for_update")
	}

	cacheKey := fmt.Sprintf("user:%s", id)
	d.redis0.Del(ctx, cacheKey)

	return nil
}

func (d *userRepository) DeleteUser(ctx context.Context, id string) error {
	query, _ := d.queryLoader.Get("DeleteUser")

	result, err := d.sql0.ExecContext(ctx, query, id)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("failed_to_delete_user")
		return x.WrapWithCode(err, x.CodeSQLDelete, "failed_to_delete_user")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("id", id).Msg("user_not_found_for_deletion")
		return x.WrapWithCode(err, x.CodeSQLEmptyRow, "user_not_found_for_deletion")
	}

	cacheKey := fmt.Sprintf("user:%s", id)
	d.redis0.Del(ctx, cacheKey)

	return nil
}
