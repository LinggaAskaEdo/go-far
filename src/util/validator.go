package util

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func Validator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("custom", customValidator)
	}
}

func customValidator(fl validator.FieldLevel) bool {
	return true
}
