package validator

import (
	"encoding/json"
	"fmt"
	"testing"

	"go-far/internal/model/entity"

	"github.com/go-playground/validator/v10"
)

type TestRequest struct {
	Name string      `json:"name" validate:"required"`
	Role entity.Role `json:"role" validate:"required,role_valid"`
}

func TestValidator(t *testing.T) {
	jsonData := `{"name":"Test","role":"userx"}`
	var req TestRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		fmt.Println("Unmarshal error:", err)
		return
	}

	validate := validator.New()

	// Register custom validation
	err = validate.RegisterValidation("role_valid", func(fl validator.FieldLevel) bool {
		role := fl.Field().Interface().(entity.Role)
		return role == entity.RoleAdmin || role == entity.RoleUser || role == entity.RoleGuest
	})
	if err != nil {
		fmt.Println("Register error:", err)
		return
	}

	// Validate struct
	err = validate.Struct(req)
	if err != nil {
		fmt.Println("Validation failed:", err) // will print error for role "userx"
	} else {
		fmt.Println("Validation passed")
	}
}
