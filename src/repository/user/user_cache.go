package user

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"go-far/src/model/dto"
	"go-far/src/model/entity"
	x "go-far/src/model/errors"

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
		return x.WrapWithCode(err, x.CodeCacheMarshal, "set_cache_find_all_user_marshal")
	}

	encJSON = snappy.Encode(encJSON, rawJSON)

	if hsetErr := d.redis0.HSet(ctx, userByParamHashKey, cacheKey, encJSON).Err(); hsetErr != nil {
		return x.WrapWithCode(hsetErr, x.CodeCacheSetHashKey, "set_cache_find_all_user")
	}

	if expireErr := d.redis0.Expire(ctx, userByParamHashKey, durationUserExpiration).Err(); expireErr != nil {
		return x.WrapWithCode(expireErr, x.CodeCacheSetExpiration, "set_cache_find_all_user_expiration")
	}

	rawJSON, err = json.Marshal(pagination)
	if err != nil {
		return x.WrapWithCode(err, x.CodeCacheMarshal, "set_cache_find_all_user_pagination_marshal")
	}

	encJSON = []byte{}
	encJSON = snappy.Encode(encJSON, rawJSON)

	if err := d.redis0.HSet(ctx, userPaginationHashKey, cacheKey, encJSON).Err(); err != nil {
		return x.WrapWithCode(err, x.CodeCacheSetHashKey, "set_cache_find_all_user_pagination")
	}

	if err := d.redis0.Expire(ctx, userPaginationHashKey, durationUserExpiration).Err(); err != nil {
		return x.WrapWithCode(err, x.CodeCacheSetExpiration, "set_cache_find_all_user_pagination_expiration")
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
	if err == redis.Nil {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheGetHashKey, "get_cache_find_all_user")
	}

	var decJSON []byte
	decJSON, err = snappy.Decode(decJSON, resultRaw)
	if err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheDecode, "get_cache_find_all_user")
	}

	if unmarshallErr := json.Unmarshal(decJSON, &results); unmarshallErr != nil {
		return nil, nil, x.WrapWithCode(unmarshallErr, x.CodeCacheUnmarshal, "get_cache_find_all_user")
	}

	paginationRaw, err := d.redis0.HGet(ctx, userPaginationHashKey, cacheKey).Bytes()
	if err == redis.Nil {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheGetHashKey, "get_cache_find_all_user_pagination")
	}

	decJSON = []byte{}
	decJSON, err = snappy.Decode(decJSON, paginationRaw)
	if err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheDecode, "get_cache_find_all_user_pagination")
	}

	if err := json.Unmarshal(decJSON, &pagination); err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheUnmarshal, "get_cache_find_all_user_pagination")
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
	var result string
	for _, k := range keys {
		if k != "" {
			result += k + "|"
		}
	}
	return result
}
