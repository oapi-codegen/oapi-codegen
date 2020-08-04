package codegen

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

const (
	extPropGoType = "x-go-type"
)

func extTypeName(extPropValue interface{}) (string, error) {
	raw, ok := extPropValue.(json.RawMessage)
	if !ok {
		return "", fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	var name string
	if err := json.Unmarshal(raw, &name); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal json")
	}

	return name, nil
}
