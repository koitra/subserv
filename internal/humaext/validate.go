package humaext

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
)

func ValidateToHuma(err error) (error, bool) {
	errs, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return err, false
	}

	model := huma.ErrorModel{
		Status: http.StatusUnprocessableEntity,
		Detail: "validation error",
		Errors: make([]*huma.ErrorDetail, len(errs)),
	}
	for i, err := range errs {
		model.Errors[i] = &huma.ErrorDetail{
			Message:  "invalid value",
			Location: err.Field(),
			Value:    err.Value(),
		}
	}

	return &model, true
}
