package user

import (
	"context"

	"go-far/src/model/dto"
	"go-far/src/model/entity"
	x "go-far/src/model/errors"

	"golang.org/x/crypto/bcrypt"
)

func (s *userService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*entity.User, error) {
	return s.registerUser(ctx, req.Name, req.Email, req.Password, req.Age, entity.Role(req.Role))
}

func (s *userService) RegisterUser(ctx context.Context, req dto.RegisterRequest) (*entity.User, error) {
	return s.registerUser(ctx, req.Name, req.Email, req.Password, req.Age, entity.Role(req.Role))
}

func (s *userService) registerUser(ctx context.Context, name, email, password string, age int, role entity.Role) (*entity.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, x.WrapWithCode(err, x.CodeHTTPInternalServerError, "failed to hash password")
	}

	user := &entity.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Age:      age,
		Role:     role,
	}

	user, err = s.userRepository.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (*entity.User, error) {
	user, err := s.userRepository.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, x.NewWithCode(x.CodeHTTPUnauthorized, "Invalid credentials")
	}

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*entity.User, error) {
	return s.userRepository.FindByID(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, cacheControl dto.CacheControl, filter *dto.UserFilter) (*[]entity.User, *dto.Pagination, error) {
	return s.userRepository.FindAll(ctx, cacheControl, filter)
}

func (s *userService) UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (*entity.User, error) {
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

	if req.Role != "" {
		existingUser.Role = entity.Role(req.Role)
	}

	if req.IsActive != nil {
		existingUser.IsActive = *req.IsActive
	}

	if err := s.userRepository.Update(ctx, existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepository.Delete(ctx, id)
}
