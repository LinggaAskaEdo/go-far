package middleware

import (
	"regexp"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

var (
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#\$%\^&\*]{8,}$`)
	onceValidator = &sync.Once{}
	validatorInst *validator.Validate
)

// InitValidator initializes custom validators
func InitValidator(log zerolog.Logger) *validator.Validate {
	onceValidator.Do(func() {
		v := validator.New()
		err := v.RegisterValidation("password", passwordValidator)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to load custom password validator")
		} else {
			log.Debug().Msg("Custom password validator loaded successfully")
		}
		validatorInst = v
	})

	return validatorInst
}

func passwordValidator(fl validator.FieldLevel) bool {
	return passwordRegex.MatchString(fl.Field().String())
}
