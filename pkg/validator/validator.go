package validation

import (
	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(s interface{}) error
}

type validatorWrapper struct {
	validator *validator.Validate
}

func NewValidator() Validator {
	validate := validator.New()

	return &validatorWrapper{validator: validate}
}

func (v *validatorWrapper) Validate(s interface{}) error {
	return v.validator.Struct(s)
}
