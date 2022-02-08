package echo_reponse_helper

import (
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type Validator struct {
}

func (v Validator) Validate(i interface{}) error {
	return nil
}

func TestGetResponseHelper(t *testing.T) {
	e := echo.New()
	e.Validator = Validator{}

	RegisterHandlers(e, new(Api))

	result := testutil.NewRequest().
		Get("/things").
		Go(t, e)

	var things []ThingWithID

	if err := result.UnmarshalJsonToObject(&things); err != nil {
		t.Error(err)
	}

	assert.Equal(t, "thing1", things[0].Name)
}

func TestPostResponseHelper(t *testing.T) {
	e := echo.New()
	e.Validator = Validator{}

	excpectedName := "thingRandom"

	RegisterHandlers(e, new(Api))

	result := testutil.NewRequest().
		Post("/things").
		WithJsonBody(AddThingJSONRequestBody{
			Name: excpectedName,
		}).
		Go(t, e)

	var things []ThingWithID

	if err := result.UnmarshalJsonToObject(&things); err != nil {
		t.Error(err)
	}

	assert.Equal(t, excpectedName, things[0].Name)
}
