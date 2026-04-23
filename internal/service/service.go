package service

import (
	"go-far/internal/repository"
	"go-far/internal/service/car"
	"go-far/internal/service/user"
)

type Service struct {
	User user.UserServiceItf
	Car  car.CarServiceItf
}

func InitService(repo *repository.Repository) *Service {
	return &Service{
		User: user.InitUserService(
			repo.User,
		),
		Car: car.InitCarService(
			repo.Car,
		),
	}
}
