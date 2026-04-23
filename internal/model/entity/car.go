package entity

import (
	"time"

	"github.com/google/uuid"
)

type Car struct {
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
	ID           string    `db:"id" json:"id"`
	Brand        string    `db:"brand" json:"brand"`
	Model        string    `db:"model" json:"model"`
	Color        string    `db:"color" json:"color"`
	LicensePlate string    `db:"license_plate" json:"license_plate"`
	Year         int       `db:"year" json:"year"`
	IsAvailable  bool      `db:"is_available" json:"is_available"`
}

type CarWithOwner struct {
	OwnerName  string `db:"owner_name" json:"owner_name"`
	OwnerEmail string `db:"owner_email" json:"owner_email"`
	Car
}

type UserCar struct {
	UserID uuid.UUID `json:"user_id"`
	CarID  uuid.UUID `json:"car_id"`
}
