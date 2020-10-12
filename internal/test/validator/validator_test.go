package validator_test

import (
	"testing"

	types "github.com/deepmap/oapi-codegen/internal/test/validator"

	"github.com/stretchr/testify/assert"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	translator ut.Translator
	validate   *validator.Validate
)

func init() {
	en := en.New()
	uni := ut.New(en, en)

	var ok bool
	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	if translator, ok = uni.GetTranslator("en"); !ok {
		panic("language not found")
	}
	validate = validator.New()
	if err := en_translations.RegisterDefaultTranslations(validate, translator); err != nil {
		panic("err")
	}
}

func TestStructA(t *testing.T) {
	a := types.StructA{}

	err := validate.Struct(a)
	if assert.Error(t, err) {
		typedErr := err.(validator.ValidationErrors)
		assert.Equal(t, "Key: 'StructA.RangeInt' Error:Field validation for 'RangeInt' failed on the 'gte' tag\nKey: 'StructA.RequiredString' Error:Field validation for 'RequiredString' failed on the 'required' tag", typedErr.Error())
		assert.Equal(t, validator.ValidationErrorsTranslations{
			"StructA.RangeInt":       "RangeInt must be 3 or greater",
			"StructA.RequiredString": "RequiredString is a required field",
		}, typedErr.Translate(translator))
	}

	var i int64 = 50
	a.RangeInt = &i
	a.RequiredString = "hello"
	//
	err = validate.Struct(a)
	if assert.Error(t, err) {
		typedErr := err.(validator.ValidationErrors)
		assert.Equal(t, validator.ValidationErrorsTranslations{
			"StructA.RangeInt": "RangeInt must be 42 or less",
		}, typedErr.Translate(translator))
	}
}

func TestStructB(t *testing.T) {
	b := types.StructB{}

	err := validate.Struct(b)
	if assert.Error(t, err) {
		typedErr := err.(validator.ValidationErrors)
		assert.Equal(t, validator.ValidationErrorsTranslations{
			"StructB.ListItem": "ListItem must contain at least 1 item",
		}, typedErr.Translate(translator))
	}
}

func TestStructC(t *testing.T) {
	c := types.StructC{
		Color: "orange",
	}

	err := validate.Struct(c)
	if assert.Error(t, err) {
		typedErr := err.(validator.ValidationErrors)
		assert.Equal(t, validator.ValidationErrorsTranslations{
			"StructC.Color": "Color must be one of [black white]",
		}, typedErr.Translate(translator))
	}
}