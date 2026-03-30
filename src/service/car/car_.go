package car

import (
	"context"

	"go-far/src/domain"
	"go-far/src/dto"

	"github.com/google/uuid"
)

func (s *carService) CreateCar(ctx context.Context, req dto.CreateCarRequest) (*domain.Car, error) {
	car := &domain.Car{
		UserID:       req.UserID.String(),
		Brand:        req.Brand,
		Model:        req.Model,
		Year:         req.Year,
		Color:        req.Color,
		LicensePlate: req.LicensePlate,
		IsAvailable:  true,
	}

	if err := s.carRepository.Create(ctx, car); err != nil {
		return nil, err
	}

	return car, nil
}

func (s *carService) CreateBulkCars(ctx context.Context, req dto.BulkCreateCarsRequest) ([]*domain.Car, error) {
	cars := make([]*domain.Car, 0, len(req.Cars))

	for _, carReq := range req.Cars {
		car := &domain.Car{
			UserID:       req.UserID.String(),
			Brand:        carReq.Brand,
			Model:        carReq.Model,
			Year:         carReq.Year,
			Color:        carReq.Color,
			LicensePlate: carReq.LicensePlate,
			IsAvailable:  true,
		}
		cars = append(cars, car)
	}

	if err := s.carRepository.CreateBulk(ctx, cars); err != nil {
		return nil, err
	}

	return cars, nil
}

func (s *carService) GetCar(ctx context.Context, id uuid.UUID) (*domain.Car, error) {
	return s.carRepository.FindByID(ctx, id)
}

func (s *carService) GetCarWithOwner(ctx context.Context, id uuid.UUID) (*domain.CarWithOwner, error) {
	return s.carRepository.FindByIDWithOwner(ctx, id)
}

func (s *carService) ListCarsByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Car, error) {
	return s.carRepository.FindByUserID(ctx, userID)
}

func (s *carService) CountCarsByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.carRepository.CountByUserID(ctx, userID)
}

func (s *carService) UpdateCar(ctx context.Context, id uuid.UUID, req dto.UpdateCarRequest) (*domain.Car, error) {
	existingCar, err := s.carRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Brand != "" {
		existingCar.Brand = req.Brand
	}

	if req.Model != "" {
		existingCar.Model = req.Model
	}

	if req.Year > 0 {
		existingCar.Year = req.Year
	}

	if req.Color != "" {
		existingCar.Color = req.Color
	}

	if req.LicensePlate != "" {
		existingCar.LicensePlate = req.LicensePlate
	}

	if req.IsAvailable != nil {
		existingCar.IsAvailable = *req.IsAvailable
	}

	if err := s.carRepository.Update(ctx, id, existingCar); err != nil {
		return nil, err
	}

	return s.carRepository.FindByID(ctx, id)
}

func (s *carService) DeleteCar(ctx context.Context, id uuid.UUID) error {
	return s.carRepository.Delete(ctx, id)
}

func (s *carService) TransferCarOwnership(ctx context.Context, carID, newUserID uuid.UUID) error {
	return s.carRepository.TransferOwnership(ctx, carID, newUserID)
}

func (s *carService) BulkUpdateAvailability(ctx context.Context, req dto.BulkUpdateAvailabilityRequest) error {
	return s.carRepository.BulkUpdateAvailability(ctx, req.CarIDs, req.IsAvailable)
}
