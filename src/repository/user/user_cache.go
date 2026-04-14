package user

import (
	"context"
	"encoding/json"
	"time"

	"go-far/src/model/dto"
	"go-far/src/model/entity"
	x "go-far/src/model/errors"

	"github.com/golang/snappy"
	"github.com/redis/go-redis/v9"
)

const (
	userByParamHashKey           string = "user:param"
	userPaginationByParamHashKey string = "user:pagination"
	durationUserExpiration              = 5 * time.Minute
)

func (d *userRepository) setCacheFindAllUser(ctx context.Context, filter dto.UserFilter, result *[]entity.User, pagination *dto.Pagination) error {
	var encJSON []byte

	rawKey, err := json.Marshal(filter)
	if err != nil {
		return x.WrapWithCode(err, x.CodeCacheMarshal, "set_cache_find_all_user_marshal")
	}

	field := string(rawKey)

	rawJSON, err := json.Marshal(result)
	if err != nil {
		return x.WrapWithCode(err, x.CodeCacheMarshal, "set_cache_find_all_user_marshal")
	}

	encJSON = snappy.Encode(encJSON, rawJSON)

	if err := d.redis0.HSet(ctx, userByParamHashKey, field, encJSON).Err(); err != nil {
		return x.WrapWithCode(err, x.CodeCacheSetHashKey, "set_cache_find_all_user")
	}

	if err := d.redis0.Expire(ctx, userByParamHashKey, durationUserExpiration).Err(); err != nil {
		return x.WrapWithCode(err, x.CodeCacheSetExpiration, "set_cache_find_all_user_expiration")
	}

	rawJSON, err = json.Marshal(pagination)
	if err != nil {
		return x.WrapWithCode(err, x.CodeCacheMarshal, "set_cache_find_all_user_pagination_marshal")
	}

	encJSON = []byte{}
	encJSON = snappy.Encode(encJSON, rawJSON)

	if err := d.redis0.HSet(ctx, userPaginationByParamHashKey, field, encJSON).Err(); err != nil {
		return x.WrapWithCode(err, x.CodeCacheSetHashKey, "set_cache_find_all_user_pagination")
	}

	if err := d.redis0.Expire(ctx, userPaginationByParamHashKey, durationUserExpiration).Err(); err != nil {
		return x.WrapWithCode(err, x.CodeCacheSetExpiration, "set_cache_find_all_user_pagination_expiration")
	}

	return nil
}

func (d *userRepository) getCacheFindAllUser(ctx context.Context, filter dto.UserFilter) (*[]entity.User, *dto.Pagination, error) {
	var (
		results    []entity.User
		pagination dto.Pagination
	)

	// serialize query param to string
	rawKey, err := json.Marshal(filter)
	if err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheMarshal, "get_cache_find_all_user_marshal")
	}

	field := string(rawKey)

	// fetch transaction
	resultRaw, err := d.redis0.HGet(ctx, userByParamHashKey, field).Bytes()
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

	if err := json.Unmarshal(decJSON, &results); err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheUnmarshal, "get_cache_find_all_user")
	}

	// fetch pagination
	paginationRaw, err := d.redis0.HGet(ctx, userPaginationByParamHashKey, field).Bytes()
	if err == redis.Nil {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheGetHashKey, "get_cache_find_all_user_pagination")
	}

	// decode pagination (encoded json)
	decJSON = []byte{}
	decJSON, err = snappy.Decode(decJSON, paginationRaw)
	if err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheDecode, "get_cache_find_all_user_pagination")
	}

	// unmarshaling returned byte
	if err := json.Unmarshal(decJSON, &pagination); err != nil {
		return nil, nil, x.WrapWithCode(err, x.CodeCacheUnmarshal, "get_cache_find_all_user_pagination")
	}

	return &results, &pagination, nil
}
