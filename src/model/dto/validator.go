package dto

import (
	"strings"

	"go-far/src/model/errors"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateRequest validates a request DTO and returns an error if validation fails.
func ValidateRequest(req interface{}) error {
	if err := validate.Struct(req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var messages []string
			for _, fe := range validationErrors {
				messages = append(messages, formatValidationError(fe))
			}
			// Return error with validator error code
			return errors.NewWithCode(errors.CodeHTTPValidatorError, strings.Join(messages, "; "))
		}
	}
	return nil
}

// formatValidationError returns a human-readable validation error message.
func formatValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email address"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "gte":
		return fe.Field() + " must be >= " + fe.Param()
	case "lte":
		return fe.Field() + " must be <= " + fe.Param()
	case "oneof":
		return fe.Field() + " must be one of: " + fe.Param()
	default:
		return fe.Field() + " failed " + fe.Tag() + " validation"
	}
}
