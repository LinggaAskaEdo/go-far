package user

import (
	"context"
	"encoding/json"
	"errors"

	"go-far/internal/model/dto"
	"go-far/internal/model/entity"
	appErr "go-far/internal/model/errors"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func (d *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	tx, err := d.sql0.Begin(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("tx_create_user")
		return user, appErr.Wrap(err, "tx_create_user")
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				zerolog.Ctx(ctx).Error().Err(rollbackErr).Msg("rollback_create_user")
			}
		}
	}()

	tx, user, err = d.createSQLUser(ctx, tx, user)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("sql_create_user")
		return nil, appErr.Wrap(err, "sql_create_user")
	}

	if err = tx.Commit(ctx); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("commit_create_user")
		return nil, appErr.Wrap(err, "commit_create_user")
	}

	return user, nil
}

func (d *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	cacheKey := "user:" + id
	cached, err := d.redis0.Get(ctx, cacheKey).Result()
	if err == nil {
		var user entity.User

		if unmarshallErr := json.Unmarshal([]byte(cached), &user); unmarshallErr == nil {
			zerolog.Ctx(ctx).Debug().Str("id", id).Msg("data_found_in_cache")
			return &user, nil
		}
	}

	user, err := d.findUserSQLByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(user)
	d.redis0.Set(ctx, cacheKey, data, d.cacheTTL)

	return user, nil
}

func (d *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := d.findUserSQLByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (d *userRepository) FindAll(ctx context.Context, cacheControl dto.CacheControl, filter *dto.UserFilter) (*[]entity.User, *dto.Pagination, error) {
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
	if errors.Is(err, redis.Nil) {
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

func (d *userRepository) Update(ctx context.Context, user *entity.User) error {
	return d.updateSQLUser(ctx, user)
}

func (d *userRepository) Delete(ctx context.Context, id string) error {
	return d.deleteSQLUser(ctx, id)
}
