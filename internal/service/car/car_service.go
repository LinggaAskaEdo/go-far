package car

import (
	"context"

	"go-far/internal/model/dto"
	"go-far/internal/model/entity"
	x "go-far/internal/model/errors"

	"github.com/google/uuid"
)

func (s *carService) CreateCar(ctx context.Context, req dto.CreateCarRequest, ownerUserID string) (*entity.Car, error) {
	car := &entity.Car{
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

	carUUID, err := uuid.Parse(car.ID)
	if err != nil {
		return nil, err
	}

	userUUID, err := uuid.Parse(ownerUserID)
	if err != nil {
		return nil, err
	}

	if err := s.carRepository.AssignCarToUser(ctx, userUUID, carUUID); err != nil {
		return nil, err
	}

	return car, nil
}

func (s *carService) CreateBulkCars(ctx context.Context, req dto.BulkCreateCarsRequest, ownerUserID string) ([]*entity.Car, error) {
	cars := make([]*entity.Car, 0, len(req.Cars))
	carIDs := make([]uuid.UUID, 0, len(req.Cars))

	for _, carReq := range req.Cars {
		car := &entity.Car{
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

	for _, car := range cars {
		carID, err := uuid.Parse(car.ID)
		if err != nil {
			return nil, err
		}
		carIDs = append(carIDs, carID)
	}

	// Assign all cars to user via junction table
	userUUID, err := uuid.Parse(ownerUserID)
	if err != nil {
		return nil, err
	}

	if err := s.carRepository.AssignCarsToUserBulk(ctx, userUUID, carIDs); err != nil {
		return nil, err
	}

	return cars, nil
}

func (s *carService) GetCar(ctx context.Context, id uuid.UUID) (*entity.Car, error) {
	return s.carRepository.FindByID(ctx, id)
}

func (s *carService) GetCarWithOwner(ctx context.Context, id uuid.UUID) (*entity.CarWithOwner, error) {
	return s.carRepository.FindByIDWithOwner(ctx, id)
}

func (s *carService) ListCarsByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Car, error) {
	return s.carRepository.FindByUserID(ctx, userID)
}

func (s *carService) CountCarsByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.carRepository.CountByUserID(ctx, userID)
}

func (s *carService) UpdateCar(ctx context.Context, id uuid.UUID, req *dto.UpdateCarRequest, userID string) (*entity.Car, error) {
	existingCar, err := s.carRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	isOwner, err := s.carRepository.IsCarOwnedByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if !isOwner {
		return nil, x.NewWithCode(x.CodeHTTPForbidden, "you do not have permission to update this car")
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

	return existingCar, nil
}

func (s *carService) DeleteCar(ctx context.Context, id uuid.UUID, userID string) error {
	isOwner, err := s.carRepository.IsCarOwnedByUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return x.NewWithCode(x.CodeHTTPForbidden, "you do not have permission to delete this car")
	}

	return s.carRepository.Delete(ctx, id)
}

func (s *carService) TransferCarOwnership(ctx context.Context, carID, newUserID uuid.UUID, userID string) error {
	isOwner, err := s.carRepository.IsCarOwnedByUser(ctx, carID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return x.NewWithCode(x.CodeHTTPForbidden, "you do not have permission to transfer this car")
	}

	return s.carRepository.TransferOwnership(ctx, carID, newUserID)
}

func (s *carService) BulkUpdateAvailability(ctx context.Context, req dto.BulkUpdateAvailabilityRequest, userID string) error {
	ownershipMap, err := s.carRepository.AreCarsOwnedByUser(ctx, req.CarIDs, userID)
	if err != nil {
		return err
	}

	for carID, isOwned := range ownershipMap {
		if !isOwned {
			return x.NewWithCode(x.CodeHTTPForbidden, "you do not have permission to update car: "+carID.String())
		}
	}

	return s.carRepository.BulkUpdateAvailability(ctx, req.CarIDs, req.IsAvailable)
}
