package user

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"go-far/internal/model/dto"
	"go-far/internal/model/entity"
	appErr "go-far/internal/model/errors"

	"github.com/golang/snappy"
	"github.com/redis/go-redis/v9"
)

const (
	userByParamHashKey     string        = "user:param"
	userPaginationHashKey  string        = "user:pagination"
	durationUserExpiration time.Duration = 5 * time.Minute
)

func (d *userRepository) setCacheFindAllUser(ctx context.Context, filter *dto.UserFilter, result *[]entity.User, pagination *dto.Pagination) error {
	cacheKey := generateCacheKey(filter)

	var encJSON []byte

	rawJSON, err := json.Marshal(result)
	if err != nil {
		return appErr.WrapWithCode(err, appErr.CodeCacheMarshal, "set_cache_find_all_user_marshal")
	}

	encJSON = snappy.Encode(encJSON, rawJSON)

	if hsetErr := d.redis0.HSet(ctx, userByParamHashKey, cacheKey, encJSON).Err(); hsetErr != nil {
		return appErr.WrapWithCode(hsetErr, appErr.CodeCacheSetHashKey, "set_cache_find_all_user")
	}

	if expireErr := d.redis0.Expire(ctx, userByParamHashKey, durationUserExpiration).Err(); expireErr != nil {
		return appErr.WrapWithCode(expireErr, appErr.CodeCacheSetExpiration, "set_cache_find_all_user_expiration")
	}

	rawJSON, err = json.Marshal(pagination)
	if err != nil {
		return appErr.WrapWithCode(err, appErr.CodeCacheMarshal, "set_cache_find_all_user_pagination_marshal")
	}

	encJSON = []byte{}
	encJSON = snappy.Encode(encJSON, rawJSON)

	if err := d.redis0.HSet(ctx, userPaginationHashKey, cacheKey, encJSON).Err(); err != nil {
		return appErr.WrapWithCode(err, appErr.CodeCacheSetHashKey, "set_cache_find_all_user_pagination")
	}

	if err := d.redis0.Expire(ctx, userPaginationHashKey, durationUserExpiration).Err(); err != nil {
		return appErr.WrapWithCode(err, appErr.CodeCacheSetExpiration, "set_cache_find_all_user_pagination_expiration")
	}

	return nil
}

func (d *userRepository) getCacheFindAllUser(ctx context.Context, filter *dto.UserFilter) (*[]entity.User, *dto.Pagination, error) {
	var (
		results    []entity.User
		pagination dto.Pagination
	)

	cacheKey := generateCacheKey(filter)

	resultRaw, err := d.redis0.HGet(ctx, userByParamHashKey, cacheKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, appErr.WrapWithCode(err, appErr.CodeCacheGetHashKey, "get_cache_find_all_user")
	}

	var decJSON []byte
	decJSON, err = snappy.Decode(decJSON, resultRaw)
	if err != nil {
		return nil, nil, appErr.WrapWithCode(err, appErr.CodeCacheDecode, "get_cache_find_all_user")
	}

	if unmarshallErr := json.Unmarshal(decJSON, &results); unmarshallErr != nil {
		return nil, nil, appErr.WrapWithCode(unmarshallErr, appErr.CodeCacheUnmarshal, "get_cache_find_all_user")
	}

	paginationRaw, err := d.redis0.HGet(ctx, userPaginationHashKey, cacheKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, appErr.WrapWithCode(err, appErr.CodeCacheGetHashKey, "get_cache_find_all_user_pagination")
	}

	decJSON = []byte{}
	decJSON, err = snappy.Decode(decJSON, paginationRaw)
	if err != nil {
		return nil, nil, appErr.WrapWithCode(err, appErr.CodeCacheDecode, "get_cache_find_all_user_pagination")
	}

	if err := json.Unmarshal(decJSON, &pagination); err != nil {
		return nil, nil, appErr.WrapWithCode(err, appErr.CodeCacheUnmarshal, "get_cache_find_all_user_pagination")
	}

	return &results, &pagination, nil
}

func generateCacheKey(filter *dto.UserFilter) string {
	keys := []string{
		"id:" + filter.ID,
		"name:" + filter.Name,
		"email:" + filter.Email,
	}
	if filter.MinAge > 0 {
		keys = append(keys, "min_age:0")
	}
	if filter.MaxAge > 0 {
		keys = append(keys, "max_age:0")
	}
	if filter.Page > 0 {
		keys = append(keys, "page:0")
	}
	if filter.PageSize > 0 {
		keys = append(keys, "page_size:0")
	}
	if filter.SortBy != "" {
		keys = append(keys, "sort_by:"+filter.SortBy)
	}
	if filter.SortDir != "" {
		keys = append(keys, "sort_dir:"+filter.SortDir)
	}

	sort.Strings(keys)
	return hashStrings(keys)
}

func hashStrings(keys []string) string {
	nonEmpty := make([]string, 0, len(keys))
	for _, k := range keys {
		if k != "" {
			nonEmpty = append(nonEmpty, k)
		}
	}

	return strings.Join(nonEmpty, "|")
}
