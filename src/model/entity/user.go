package entity

import "time"

// Role represents a user role in the system
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"
)

func (r Role) String() string {
	return string(r)
}

func ParseRole(s string) Role {
	switch Role(s) {
	case RoleAdmin, RoleUser, RoleGuest:
		return Role(s)
	default:
		return RoleUser
	}
}

type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	Password  string    `db:"password" json:"-"`
	Age       int       `db:"age" json:"age"`
	Role      Role      `db:"role" json:"role"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type UserWithCars struct {
	User
	Cars     []Car `json:"cars,omitempty"`
	CarCount int   `json:"car_count"`
}
