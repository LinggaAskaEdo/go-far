package car

import (
	"context"

	"go-far/src/model/dto"
	"go-far/src/model/entity"
	"go-far/src/repository/car"

	"github.com/google/uuid"
)

type CarServiceItf interface {
	CreateCar(ctx context.Context, req dto.CreateCarRequest, ownerUserID string) (*entity.Car, error)
	CreateBulkCars(ctx context.Context, req dto.BulkCreateCarsRequest, ownerUserID string) ([]*entity.Car, error)
	GetCar(ctx context.Context, id uuid.UUID) (*entity.Car, error)
	GetCarWithOwner(ctx context.Context, id uuid.UUID) (*entity.CarWithOwner, error)
	ListCarsByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Car, error)
	CountCarsByUser(ctx context.Context, userID uuid.UUID) (int, error)
	UpdateCar(ctx context.Context, id uuid.UUID, req *dto.UpdateCarRequest, userID string) (*entity.Car, error)
	DeleteCar(ctx context.Context, id uuid.UUID, userID string) error
	TransferCarOwnership(ctx context.Context, carID, newUserID uuid.UUID, userID string) error
	BulkUpdateAvailability(ctx context.Context, req dto.BulkUpdateAvailabilityRequest, userID string) error
}

type carService struct {
	carRepository car.CarRepositoryItf
}

func InitCarService(carRepository car.CarRepositoryItf) CarServiceItf {
	return &carService{
		carRepository: carRepository,
	}
}
