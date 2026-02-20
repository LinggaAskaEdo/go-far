package user

import (
	"context"

	"go-far/src/domain"
	"go-far/src/dto"
)

func (s *userService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*domain.User, error) {
	user := &domain.User{
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	if _, err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (domain.User, error) {
	return s.userRepository.FindByID(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, cacheControl dto.CacheControl, filter dto.UserFilter) ([]domain.User, dto.Pagination, error) {
	return s.userRepository.FindAll(ctx, cacheControl, filter)
}

func (s *userService) UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (domain.User, error) {
	existingUser, err := s.userRepository.FindByID(ctx, id)
	if err != nil {
		return existingUser, err
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
		return existingUser, err
	}

	return s.userRepository.FindByID(ctx, id)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepository.Delete(ctx, id)
}
