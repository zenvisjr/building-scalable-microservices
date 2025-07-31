package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(i interface{}) error {
	err := validate.Struct(i)
	if err == nil {
		return nil
	}
	var sb strings.Builder
	for _, e := range err.(validator.ValidationErrors) {
		sb.WriteString(fmt.Sprintf("%s: %s; ", e.Field(), e.Tag()))
	}
	return errors.New(sb.String())
}
