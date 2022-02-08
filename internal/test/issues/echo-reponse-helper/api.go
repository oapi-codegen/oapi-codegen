package echo_reponse_helper

type Api struct {
}

func (a Api) ListThings(ctx ListThingsContext) error {
	return ctx.JSON200([]ThingWithID{
		{
			Thing: Thing{Name: "thing1"},
			Id:    1,
		},
		{
			Thing: Thing{Name: "thing2"},
			Id:    2,
		},
	})
}

func (a Api) AddThing(ctx AddThingContext) error {
	body, err := ctx.BindJSON()
	if err != nil {
		return ctx.JSON400(Error{
			Message: err.Error(),
			Code:    400,
		})
	}

	return ctx.JSON201([]ThingWithID{
		{
			Thing: Thing(*body),
			Id:    12,
		},
	})
}
