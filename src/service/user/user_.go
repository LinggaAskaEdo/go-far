package user

import (
	"context"
	"go-far/src/domain"

	"github.com/rs/zerolog"
)

func (s *userService) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	user := &domain.User{
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepository.FindByID(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, filter domain.UserFilter) ([]*domain.User, int64, error) {
	zerolog.Ctx(ctx).Debug().Interface("filter", filter).Msg("Listing users with filter")
	return s.userRepository.FindAll(ctx, filter)
}

func (s *userService) UpdateUser(ctx context.Context, id int64, req domain.UpdateUserRequest) (*domain.User, error) {
	existingUser, err := s.userRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		existingUser.Name = req.Name
	}

	if req.Email != "" {
		existingUser.Email = req.Email
	}

	if req.Age > 0 {
		existingUser.Age = req.Age
	}

	if err := s.userRepository.Update(ctx, id, existingUser); err != nil {
		return nil, err
	}

	return s.userRepository.FindByID(ctx, id)
}

func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	return s.userRepository.Delete(ctx, id)
}
