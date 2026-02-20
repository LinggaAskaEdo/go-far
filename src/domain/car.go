package domain

import (
	"time"
)

type Car struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	Brand        string    `db:"brand" json:"brand"`
	Model        string    `db:"model" json:"model"`
	Year         int       `db:"year" json:"year"`
	Color        string    `db:"color" json:"color"`
	LicensePlate string    `db:"license_plate" json:"license_plate"`
	IsAvailable  bool      `db:"is_available" json:"is_available"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type CarWithOwner struct {
	Car
	OwnerName  string `db:"owner_name" json:"owner_name"`
	OwnerEmail string `db:"owner_email" json:"owner_email"`
}
