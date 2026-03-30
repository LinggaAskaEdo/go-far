package service

import (
	"go-far/src/repository"
	"go-far/src/service/car"
	"go-far/src/service/user"
)

type Service struct {
	User user.UserServiceItf
	Car  car.CarServiceItf
}

func InitService(repository *repository.Repository) *Service {
	return &Service{
		User: user.InitUserService(
			repository.User,
		),
		Car: car.InitCarService(
			repository.Car,
		),
	}
}
