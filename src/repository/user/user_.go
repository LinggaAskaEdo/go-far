package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go-far/src/domain"
	exception "go-far/src/errors"

	"github.com/rs/zerolog"
)

func (d *userRepository) Create(ctx context.Context, user *domain.User) error {
	query, _ := d.queryLoader.Get("CreateUser")

	now := time.Now()
	err := d.db.QueryRowContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Age,
		now,
		now,
	).Scan(&user.ID)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to create user")
		return exception.InternalServerError("Failed to create user")
	}

	user.CreatedAt = now
	user.UpdatedAt = now

	return nil
}

func (d *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	cacheKey := fmt.Sprintf("user:%s", id)

	cached, err := d.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var user domain.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			zerolog.Ctx(ctx).Debug().Str("id", id).Msg("User found in cache")
			return &user, nil
		}
	}

	query, _ := d.queryLoader.Get("FindUserByID")

	var user domain.User
	err = d.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			zerolog.Ctx(ctx).Debug().Str("id", id).Msg("User not found")
			return nil, exception.NotFoundError("User not found")
		}

		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("Failed to find user")
		return nil, exception.InternalServerError("Failed to find user")
	}

	data, _ := json.Marshal(user)
	d.redis.Set(ctx, cacheKey, data, d.cacheTTL)

	return &user, nil
}

func (d *userRepository) FindAll(ctx context.Context, filter domain.UserFilter) ([]*domain.User, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.PageSize < 1 {
		filter.PageSize = 10
	}

	if filter.SortDir == "" {
		filter.SortDir = "ASC"
	}

	// Prepare template data
	templateData := map[string]any{
		"Name":    filter.Name,
		"Email":   filter.Email,
		"MinAge":  filter.MinAge,
		"MaxAge":  filter.MaxAge,
		"SortBy":  filter.SortBy,
		"SortDir": filter.SortDir,
		"name":    filter.Name,
		"email":   filter.Email,
		"min_age": filter.MinAge,
		"max_age": filter.MaxAge,
		"limit":   filter.PageSize,
		"offset":  (filter.Page - 1) * filter.PageSize,
	}

	// Count total
	countQuery, countArgs, err := d.queryLoader.ExecuteTemplate("CountUsersBase", templateData)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build count query")
		return nil, 0, exception.InternalServerError("Failed to build count query")
	}

	var total int64
	err = d.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to count users")
		return nil, 0, exception.InternalServerError("Failed to count users")
	}

	zerolog.Ctx(ctx).Debug().Int64("total", total).Msg("Total users found")

	// Get users
	query, args, err := d.queryLoader.ExecuteTemplate("FindAllUsersBase", templateData)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build query")
		return nil, 0, exception.InternalServerError("Failed to build query")
	}

	var users []*domain.User
	err = d.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to find users")
		return nil, 0, exception.InternalServerError("Failed to find users")
	}

	return users, total, nil
}

func (d *userRepository) Update(ctx context.Context, id string, user *domain.User) error {
	query, _ := d.queryLoader.Get("UpdateUser")

	result, err := d.db.ExecContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Age,
		time.Now(),
		id,
	)

	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("Failed to update user")
		return exception.InternalServerError("Failed to update user")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("id", id).Msg("User not found for update")
		return exception.NotFoundError("User not found")
	}

	cacheKey := fmt.Sprintf("user:%s", id)
	d.redis.Del(ctx, cacheKey)

	return nil
}

func (d *userRepository) Delete(ctx context.Context, id string) error {
	query, _ := d.queryLoader.Get("DeleteUser")

	result, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("Failed to delete user")
		return exception.InternalServerError("Failed to delete user")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("id", id).Msg("User not found for deletion")
		return exception.NotFoundError("User not found")
	}

	cacheKey := fmt.Sprintf("user:%s", id)
	d.redis.Del(ctx, cacheKey)

	return nil
}
