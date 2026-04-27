package entity

import "time"

// Role represents a user role in the system
type Role string

type User struct {
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	Password  string    `db:"password" json:"-"`
	Role      Role      `db:"role" json:"role"`
	Age       int       `db:"age" json:"age"`
	IsActive  bool      `db:"is_active" json:"is_active"`
}

type UserWithCars struct {
	Cars []Car `json:"cars,omitempty"`
	User
	CarCount int `json:"car_count"`
}

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
