package utils

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func FormatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				return fmt.Sprintf("%s là bắt buộc", e.Field())
			case "email":
				return fmt.Sprintf("%s không đúng mẫu", e.Field())
			case "min":
				return fmt.Sprintf("%s nhất định có ít nhất %s ký tự", e.Field(), e.Param())
			case "max":
				return fmt.Sprintf("%s nhất định có ít nhiều %s ký tự", e.Field(), e.Param())
			default:
				return fmt.Sprintf("%s không hợp lệ", e.Field())
			}
		}
	}
	return err.Error()
}
