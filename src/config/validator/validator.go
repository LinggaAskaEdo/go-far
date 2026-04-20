package validator

import (
	"go-far/src/model/entity"
	"go-far/src/model/errors"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

var (
	once sync.Once
	val  *Validator
)

type Validator struct {
	*validator.Validate
}

func InitValidator(log zerolog.Logger) {
	once.Do(func() {
		validate := validator.New()

		err := validate.RegisterValidation("role_valid", func(fl validator.FieldLevel) bool {
			role, ok := fl.Field().Interface().(entity.Role)
			if !ok {
				return false
			}

			return role == entity.RoleAdmin || role == entity.RoleUser || role == entity.RoleGuest
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to register role validation")
		}

		val = &Validator{Validate: validate}
	})
}

func ValidateRequest(req any) error {
	if val == nil {
		return errors.NewWithCode(errors.CodeHTTPValidatorError, "validator not initialized")
	}

	err := val.Struct(req)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var messages []string
			for _, fe := range validationErrors {
				messages = append(messages, formatValidationError(fe))
			}

			return errors.NewWithCode(errors.CodeHTTPValidatorError, strings.Join(messages, "; "))
		}
	}

	return nil
}

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
	case "gt":
		return fe.Field() + " must be > " + fe.Param()
	case "gte":
		return fe.Field() + " must be >= " + fe.Param()
	case "lt":
		return fe.Field() + " must be < " + fe.Param()
	case "lte":
		return fe.Field() + " must be <= " + fe.Param()
	case "oneof":
		return fe.Field() + " must be one of: " + fe.Param()
	case "role_valid":
		return fe.Field() + " must be one of: admin, user, guest"
	default:
		return fe.Field() + " failed " + fe.Tag() + " validation"
	}
}
