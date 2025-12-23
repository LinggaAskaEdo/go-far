package user

import (
	"context"
	"fmt"
	"time"

	"go-far/src/domain"
	"go-far/src/dto"
	x "go-far/src/errors"
	"go-far/src/util"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

func (d *userRepository) createSQLUser(ctx context.Context, tx *sqlx.Tx, user *domain.User) (*sqlx.Tx, *domain.User, error) {
	query, _ := d.queryLoader.Get("CreateUser")
	row := tx.QueryRowContext(ctx, query, user.Name, user.Email, user.Age).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err := row; err != nil {
		return tx, user, x.Wrap(err, "create_sql_user")
	}

	return tx, user, nil
}

func (d *userRepository) findAllSQLUser(ctx context.Context, filter dto.UserFilter) ([]domain.User, dto.Pagination, error) {
	var (
		results      []domain.User
		totalRecords int64
	)

	filter.Page = util.ValidatePage(filter.Page)
	filter.PageSize = util.ValidatePage(filter.PageSize)
	filter.SortBy = util.ValidateSortBy(filter.SortBy)
	filter.SortDir = util.ValidateSortDir(filter.SortDir)

	pagination := dto.Pagination{
		CurrentPage:     filter.Page,
		CurrentElements: 0,
		TotalPages:      0,
		TotalElements:   0,
		SortBy:          filter.SortBy,
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

	// Get users
	query, args, err := d.queryLoader.ExecuteTemplate("FindAllUsersBase", templateData)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_users_query_err")
		return nil, pagination, x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_find_users_query_err")
	}

	err = d.sql0.SelectContext(ctx, &results, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("find_users_err")
		return nil, pagination, x.WrapWithCode(err, x.CodeSQLRowScan, "find_users_err")
	}

	// Count users
	countQuery, countArgs, err := d.queryLoader.ExecuteTemplate("CountUsersBase", templateData)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("count_users_query_err")
		return nil, pagination, x.WrapWithCode(err, x.CodeSQLQueryBuild, "count_users_query_err")
	}

	err = d.sql0.GetContext(ctx, &totalRecords, countQuery, countArgs...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("count_users_err")
		return nil, pagination, x.WrapWithCode(err, x.CodeSQLRowScan, "count_users_err")
	}

	zerolog.Ctx(ctx).Debug().Int64("total", totalRecords).Msg("total_users_found")

	// Update Pagination
	totalPage := totalRecords / filter.PageSize
	if totalRecords%filter.PageSize > 0 || totalRecords == 0 {
		totalPage++
	}

	pagination.TotalPages = util.ValidatePage(totalPage)
	pagination.CurrentElements = int64(len(results))
	pagination.TotalElements = totalRecords

	return results, pagination, nil
}

func (d *userRepository) updateSQLUser(ctx context.Context, id string, user domain.User) error {
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

func (d *userRepository) deleteSQLUser(ctx context.Context, id string) error {
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
