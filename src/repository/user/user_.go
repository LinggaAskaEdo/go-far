package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go-far/src/model/dto"
	"go-far/src/model/entity"
	x "go-far/src/model/errors"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func (d *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	tx, err := d.sql0.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("tx_create_user")
		return user, x.Wrap(err, "tx_create_user")
	}

	tx, user, err = d.createSQLUser(ctx, tx, user)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			zerolog.Ctx(ctx).Error().Err(rollbackErr).Msg("rollback_create_user")
		}
		zerolog.Ctx(ctx).Error().Err(err).Msg("sql_create_user")
		return user, x.Wrap(err, "sql_create_user")
	}

	if err = tx.Commit(); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			zerolog.Ctx(ctx).Error().Err(rollbackErr).Msg("rollback_after_commit_failure")
		}
		zerolog.Ctx(ctx).Error().Err(err).Msg("commit_create_user")
		return user, x.Wrap(err, "commit_create_user")
	}

	return user, nil
}

func (d *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User

	cacheKey := fmt.Sprintf("user:%s", id)

	cached, err := d.redis0.Get(ctx, cacheKey).Result()
	if err == nil {

		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			zerolog.Ctx(ctx).Debug().Str("id", id).Msg("data_found_in_cache")
			return &user, nil
		}
	}

	query, args, err := d.queryLoader.Compile("FindUserByID", map[string]any{"ID": id})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_user_query_err")
		return nil, x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_find_user_query_err")
	}

	err = d.sql0.GetContext(ctx, &user, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			zerolog.Ctx(ctx).Debug().Str("id", id).Msg("user_not_found")
			return nil, x.WrapWithCode(err, x.CodeSQLEmptyRow, "user_not_found")
		}

		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("find_user_err")
		return nil, x.WrapWithCode(err, x.CodeSQLRowScan, "find_user_err")
	}

	data, _ := json.Marshal(user)
	d.redis0.Set(ctx, cacheKey, data, d.cacheTTL)

	return &user, nil
}

func (d *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User

	query, args, err := d.queryLoader.Compile("FindUserByEmail", map[string]any{"Email": email})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_user_by_email_query_err")
		return nil, x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_find_user_by_email_query_err")
	}

	err = d.sql0.GetContext(ctx, &user, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, x.NewWithCode(x.CodeHTTPUnauthorized, "Invalid credentials")
		}
		zerolog.Ctx(ctx).Error().Err(err).Str("email", email).Msg("find_user_by_email_err")
		return nil, x.WrapWithCode(err, x.CodeSQLRowScan, "find_user_by_email_err")
	}

	return &user, nil
}

func (d *userRepository) FindAll(ctx context.Context, cacheControl dto.CacheControl, filter dto.UserFilter) (*[]entity.User, *dto.Pagination, error) {
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

		return d.findAllSQLUser(ctx, filter)
	}

	return result, pagination, nil
}

func (d *userRepository) Update(ctx context.Context, id string, user *entity.User) error {
	return d.updateSQLUser(ctx, id, user)
}

func (d *userRepository) Delete(ctx context.Context, id string) error {
	return d.deleteSQLUser(ctx, id)
}
