package car

import (
	"context"

	"go-far/src/domain"
	"go-far/src/dto"
	"go-far/src/repository/car"

	"github.com/google/uuid"
)

type CarServiceItf interface {
	CreateCar(ctx context.Context, req dto.CreateCarRequest) (*domain.Car, error)
	CreateBulkCars(ctx context.Context, req dto.BulkCreateCarsRequest) ([]*domain.Car, error)
	GetCar(ctx context.Context, id uuid.UUID) (*domain.Car, error)
	GetCarWithOwner(ctx context.Context, id uuid.UUID) (*domain.CarWithOwner, error)
	ListCarsByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Car, error)
	CountCarsByUser(ctx context.Context, userID uuid.UUID) (int, error)
	UpdateCar(ctx context.Context, id uuid.UUID, req dto.UpdateCarRequest) (*domain.Car, error)
	DeleteCar(ctx context.Context, id uuid.UUID) error
	TransferCarOwnership(ctx context.Context, carID, newUserID uuid.UUID) error
	BulkUpdateAvailability(ctx context.Context, req dto.BulkUpdateAvailabilityRequest) error
}

type carService struct {
	carRepository car.CarRepositoryItf
}

func InitCarService(carRepository car.CarRepositoryItf) CarServiceItf {
	return &carService{
		carRepository: carRepository,
	}
}
