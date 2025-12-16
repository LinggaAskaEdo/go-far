package config

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

var passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#\$%\^&\*]{8,}$`)

func InitValidator(log zerolog.Logger) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("password", passwordValidator)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to load custom password validator")
		} else {
			log.Debug().Msg("Custom password validator loaded successfully")
		}
	}
}

func passwordValidator(fl validator.FieldLevel) bool {
	return passwordRegex.MatchString(fl.Field().String())
}
