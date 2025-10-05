package config

import (
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	// Regex for name, only a-z, A-Z, ., ', and -
	nameRegex = regexp.MustCompile(`^[a-zA-Z .'-]+$`)
)

func ValidatorInit() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		return field.Tag.Get("json")
	})

	// Register custom validation
	validate.RegisterValidation("validName", validName)

	return validate
}

func validName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return nameRegex.MatchString(value)
}
