package user

import (
	"context"

	"go-far/src/model/dto"
	"go-far/src/model/entity"
)

func (s *userService) ListUsersV2(ctx context.Context, filter *dto.UserFilterV2) (*[]entity.User, *dto.Pagination, error) {
	return s.userRepository.FindAllV2(ctx, filter)
}
