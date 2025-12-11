package user

import (
	"context"

	"go-far/src/domain"
	"go-far/src/repository/user"
)

type UserServiceItf interface {
	CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	ListUsers(ctx context.Context, filter domain.UserFilter) ([]*domain.User, int64, error)
	UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error)
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
