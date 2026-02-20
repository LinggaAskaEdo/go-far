package car

// import (
// 	"context"
// 	"errors"
// 	"go-far/src/domain"

// 	"github.com/lib/pq"
// )

// func (r *carRepository) Create(ctx context.Context, car *domain.Car) error {
// 	query := `
// 		INSERT INTO cars (user_id, brand, model, year, color, license_plate, is_available)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		RETURNING id, created_at, updated_at
// 	`

// 	err := r.getDB().QueryRowContext(
// 		ctx, query,
// 		car.UserID, car.Brand, car.Model, car.Year,
// 		car.Color, car.LicensePlate, car.IsAvailable,
// 	).Scan(&car.ID, &car.CreatedAt, &car.UpdatedAt)

// 	if err != nil {
// 		var pqErr *pq.Error
// 		if errors.As(err, &pqErr) {
// 			if pqErr.Code == "23505" {
// 				return dto.ConflictWrap(err, "License plate already exists")
// 			}
// 			if pqErr.Message == "User cannot have more than 50 cars" {
// 				return dto.BadRequestWrap(err, "User has reached maximum limit of 50 cars")
// 			}
// 		}
// 		return dto.InternalServerWrap(err, "Failed to create car")
// 	}

// 	return nil
// }
