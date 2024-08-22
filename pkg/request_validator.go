package pkg

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

var validate = validator.New()

func ValidateRequest(request interface{}) error {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	err := validate.Struct(request)

	if err != nil {
		if castedObject, ok := err.(validator.ValidationErrors); ok {
			var report string
			for _, err := range castedObject {
				switch err.Tag() {
				case "required":
					report = fmt.Sprintf("field %s is required", err.Field())
				case "email":
					report = "invalid email address"
				case "boolean":
					report = "invalid boolean"
				case "numeric":
					report = "invalid numeric/must be a number"
				case "gte", "min":
					report = fmt.Sprintf("%s value must be greater than %s", err.Field(), err.Param())
				case "lte", "max":
					report = fmt.Sprintf("%s value must be lower than %s", err.Field(), err.Param())
				case "url":
					report = fmt.Sprintf("invalid URL '%s'", err.Value())
				default:
					report = fmt.Sprintf("invalid field %s", err.Field())
				}
				if report != "" {
					break
				}
			}
			return fmt.Errorf(report)
		}
		return err
	}
	return nil
}
