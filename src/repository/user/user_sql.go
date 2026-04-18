package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-far/src/model/dto"
	"go-far/src/model/entity"
	x "go-far/src/model/errors"
	"go-far/src/util"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

var (
	allowedSortFields = map[string]string{
		"id":         "id",
		"name":       "name",
		"email":      "email",
		"age":        "age",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}

	allowedSortDirs = map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
	}
)

func sanitizeSortBy(sortBy string) string {
	normalized := normalizeString(sortBy)
	if col, ok := allowedSortFields[normalized]; ok {
		return col
	}

	return "id"
}

func sanitizeSortDir(sortDir string) string {
	normalized := normalizeString(sortDir)
	if dir, ok := allowedSortDirs[normalized]; ok {
		return dir
	}

	return "ASC"
}

func normalizeString(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return s
}

func (d *userRepository) createSQLUser(ctx context.Context, tx *sqlx.Tx, user *entity.User) (*sqlx.Tx, *entity.User, error) {
	query, args, err := d.queryLoader.Compile("CreateUser", user)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_create_user_query_err")
		return tx, user, x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_create_user_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", query).Any("args", args).Msg("compiled_query")

	row := tx.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err := row; err != nil {
		return tx, user, x.Wrap(err, "create_sql_user")
	}

	return tx, user, nil
}

func (d *userRepository) findAllSQLUser(ctx context.Context, filter dto.UserFilter) (*[]entity.User, *dto.Pagination, error) {
	var (
		results      []entity.User
		totalRecords int64
	)

	filter.Page = util.ValidatePage(filter.Page)
	filter.PageSize = util.ValidatePage(filter.PageSize)
	filter.SortBy = sanitizeSortBy(filter.SortBy)
	filter.SortDir = sanitizeSortDir(filter.SortDir)

	pagination := dto.Pagination{
		CurrentPage:     filter.Page,
		CurrentElements: 0,
		TotalPages:      0,
		TotalElements:   0,
		SortBy:          filter.SortBy,
	}

	templateData := map[string]any{
		"ID":           filter.ID,
		"Name":         filter.Name,
		"Email":        filter.Email,
		"NamePattern":  "%" + filter.Name + "%",
		"EmailPattern": "%" + filter.Email + "%",
		"MinAge":       filter.MinAge,
		"MaxAge":       filter.MaxAge,
		"Limit":        filter.PageSize,
		"Offset":       (filter.Page - 1) * filter.PageSize,
		"SortBy":       filter.SortBy,
		"SortDir":      filter.SortDir,
	}

	query, args, err := d.queryLoader.Compile("FindAllUsersBase", templateData)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_find_users_query_err")
		return nil, &pagination, x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_find_users_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", query).Any("args", args).Msg("compiled_query")

	err = d.sql0.SelectContext(ctx, &results, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("find_users_err")
		return nil, &pagination, x.WrapWithCode(err, x.CodeSQLRowScan, "find_users_err")
	}

	countQuery, countArgs, err := d.queryLoader.Compile("CountUsersBase", templateData)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("count_users_query_err")
		return nil, &pagination, x.WrapWithCode(err, x.CodeSQLQueryBuild, "count_users_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", countQuery).Any("args", countArgs).Msg("compiled_query")

	err = d.sql0.GetContext(ctx, &totalRecords, countQuery, countArgs...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("count_users_err")
		return nil, &pagination, x.WrapWithCode(err, x.CodeSQLRowScan, "count_users_err")
	}

	zerolog.Ctx(ctx).Debug().Int64("total", totalRecords).Msg("total_users_found")

	// Calculate total pages with proper handling of empty results
	var totalPage int64
	if totalRecords > 0 {
		totalPage = (totalRecords + filter.PageSize - 1) / filter.PageSize
	} else {
		totalPage = 0
	}

	pagination.TotalPages = totalPage
	pagination.CurrentElements = int64(len(results))
	pagination.TotalElements = totalRecords

	return &results, &pagination, nil
}

func (d *userRepository) updateSQLUser(ctx context.Context, id string, user *entity.User) error {
	data := map[string]any{
		"ID":        id,
		"Name":      user.Name,
		"Email":     user.Email,
		"Age":       user.Age,
		"Role":      user.Role,
		"IsActive":  user.IsActive,
		"UpdatedAt": time.Now(),
	}

	query, args, err := d.queryLoader.Compile("UpdateUser", data)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_update_user_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_update_user_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", query).Any("args", args).Msg("compiled_query")

	result, err := d.sql0.ExecContext(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("update_user_err")
		return x.WrapWithCode(err, x.CodeSQLUpdate, "update_user_err")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("failed_to_get_rows_affected")
		return x.WrapWithCode(err, x.CodeSQLUpdate, "failed_to_get_rows_affected")
	}

	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("id", id).Msg("user_not_found_for_update")
		return x.NewWithCode(x.CodeSQLEmptyRow, "user_not_found_for_update")
	}

	cacheKey := fmt.Sprintf("user:%s", id)
	d.redis0.Del(ctx, cacheKey)

	return nil
}

func (d *userRepository) deleteSQLUser(ctx context.Context, id string) error {
	query, args, err := d.queryLoader.Compile("DeleteUser", map[string]any{"ID": id})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_delete_user_query_err")
		return x.WrapWithCode(err, x.CodeSQLQueryBuild, "build_delete_user_query_err")
	}

	zerolog.Ctx(ctx).Debug().Str("query", query).Any("args", args).Msg("compiled_query")

	result, err := d.sql0.ExecContext(ctx, query, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("failed_to_delete_user")
		return x.WrapWithCode(err, x.CodeSQLDelete, "failed_to_delete_user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("id", id).Msg("failed_to_get_rows_affected")
		return x.WrapWithCode(err, x.CodeSQLDelete, "failed_to_get_rows_affected")
	}

	if rows == 0 {
		zerolog.Ctx(ctx).Debug().Str("id", id).Msg("user_not_found_for_deletion")
		return x.NewWithCode(x.CodeSQLEmptyRow, "user_not_found_for_deletion")
	}

	cacheKey := fmt.Sprintf("user:%s", id)
	d.redis0.Del(ctx, cacheKey)

	return nil
}
