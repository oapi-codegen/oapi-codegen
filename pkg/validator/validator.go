package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func NewWithPatternValidator() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("pattern", PatternValidator)
	return v
}

func PatternValidator(fl validator.FieldLevel) bool {
	p := fl.Param()
	// decode the 3 characters that would have broken the annotation-syntax
	p = strings.ReplaceAll(p, "0x22", "\"")
	p = strings.ReplaceAll(p, "0x2c", ",")
	p = strings.ReplaceAll(p, "0x60", "`")

	reg := regexp.MustCompile(p)
	return reg.MatchString(fl.Field().String())
}
