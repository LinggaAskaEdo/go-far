package user

import (
	"context"

	"go-far/src/model/dto"
	"go-far/src/model/entity"
	"go-far/src/repository/user"
)

type UserServiceItf interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*entity.User, error)
	RegisterUser(ctx context.Context, req dto.RegisterRequest) (*entity.User, error)
	Login(ctx context.Context, req dto.LoginRequest) (*entity.User, error)
	GetUser(ctx context.Context, id string) (*entity.User, error)
	ListUsers(ctx context.Context, cacheControl dto.CacheControl, filter dto.UserFilter) (*[]entity.User, *dto.Pagination, error)
	ListUsersV2(ctx context.Context, filter dto.UserFilterV2) (*[]entity.User, *dto.Pagination, error)
	UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (*entity.User, error)
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
