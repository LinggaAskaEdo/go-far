package service

import (
	"go-far/src/repository"
	"go-far/src/service/user"
)

type Service struct {
	User user.UserServiceItf
}

func InitService(repository *repository.Repository) *Service {
	return &Service{
		User: user.InitUserService(
			repository.User,
		),
	}
}
