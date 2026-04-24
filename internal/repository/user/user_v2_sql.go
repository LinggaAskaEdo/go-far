package user

import (
	"context"

	"go-far/internal/infra/query"
	"go-far/internal/model/dto"
	"go-far/internal/model/entity"
	appErr "go-far/internal/model/errors"
	"go-far/internal/util"

	"github.com/rs/zerolog"
)

func (d *userRepository) findAllSQLUserV2(ctx context.Context, filter *dto.UserFilterV2) (*[]entity.User, *dto.Pagination, error) {
	filter.Page = util.ValidatePage(filter.Page)
	filter.PageSize = util.ValidateLimit(filter.PageSize)

	pagination := dto.Pagination{
		CurrentPage:     filter.Page,
		CurrentElements: 0,
		TotalPages:      0,
		TotalElements:   0,
		SortBy:          filter.SortBy,
		SortDir:         filter.SortDir,
	}

	filter.SortBy = sanitizeSortByV2(filter.SortBy)
	filter.SortDir = sanitizeSortDirV2(filter.SortDir)

	baseQuery, _, err := d.queryLoader.Compile("FindUsersBaseV2", nil)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_users_query_err")
		return nil, &pagination, appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_users_query_err")
	}

	qb := query.NewSQLBuilder("param", "db", "", filter.Page, filter.PageSize)
	qb.AliasPrefix("-", &filter)

	queryExt, _, args, err := qb.Build()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("build_users_query_err")
		return nil, &pagination, appErr.WrapWithCode(err, appErr.CodeSQLQueryBuild, "build_users_query_err")
	}

	fullQuery := baseQuery + queryExt
	zerolog.Ctx(ctx).Debug().Str("query", util.CleanQuery(fullQuery)).Any("args", args).Msg("compiled_query")

	var results []entity.User
	rows, err := d.sql0.Query(ctx, fullQuery, args...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("find_users_err")
		return nil, &pagination, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "find_users_err")
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Age, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("scan_user_err")
			return nil, &pagination, appErr.WrapWithCode(err, appErr.CodeSQLRowScan, "scan_user_err")
		}

		results = append(results, user)
	}

	userPtr := &results
	return userPtr, &pagination, nil
}

func sanitizeSortByV2(sort string) string {
	validSortFields := map[string]bool{
		"id":         true,
		"name":       true,
		"email":      true,
		"age":        true,
		"created_at": true,
		"updated_at": true,
	}

	if sort == "" || !validSortFields[sort] {
		return "created_at"
	}

	return sort
}

func sanitizeSortDirV2(sortDir string) string {
	if sortDir == "" {
		return "ASC"
	}

	if sortDir == "desc" {
		return "DESC"
	}

	return "ASC"
}
