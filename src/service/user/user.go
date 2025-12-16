package user

import (
	"context"

	"go-far/src/domain"
	"go-far/src/dto"
	"go-far/src/repository/user"
)

type UserServiceItf interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*domain.User, error)
	GetUser(ctx context.Context, id string) (domain.User, error)
	ListUsers(ctx context.Context, cacheControl dto.CacheControl, filter dto.UserFilter) ([]domain.User, dto.Pagination, error)
	UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (domain.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type userService struct {
	userRepository user.UserRepositoryItf
}

func InitUserService(userRepository user.UserRepositoryItf) UserServiceItf {
	return &userService{
		userRepository: userRepository,
	}
}
