package user

import (
	"context"

	"go-far/internal/model/dto"
	"go-far/internal/model/entity"
)

func (d *userRepository) FindAllV2(ctx context.Context, filter *dto.UserFilterV2) (*[]entity.User, *dto.Pagination, error) {
	result, pagination, err := d.findAllSQLUserV2(ctx, filter)
	if err != nil {
		return nil, pagination, err
	}

	return result, pagination, nil
}
