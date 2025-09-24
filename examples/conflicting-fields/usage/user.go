package usage

import (
	"context"
	"errors"

	conflicting_fields "github.com/oapi-codegen/oapi-codegen/examples/conflicting-fields"
)

func UseTheThing() error {
	client, err := conflicting_fields.NewClientWithResponses(
		`http://localhost:8080/myService`,
	)
	if err != nil {
		return err
	}

	resp, err := client.GetThingWithResponse(
		context.Background(),
	)
	if err != nil {
		return err
	}

	// The `Status` should be the field on the response defined in the OAS.
	// Currently, it conflicts with the method returning the HTTPResponse.Status
	if *resp.Status != `running` {
		return errors.New(`thing isn't running`)
	}

	return nil
}
