// Package oldnullable provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package oldnullable

import (
	"encoding/json"
	"fmt"
)

// WithAdditionalProperties defines model for WithAdditionalProperties.
type WithAdditionalProperties struct {
	Nullable             *int                   `json:"Nullable"`
	Optional             *int                   `json:"Optional,omitempty"`
	ReadOnly             *int                   `json:"ReadOnly,omitempty"`
	Required             int                    `json:"Required"`
	WriteOnly            *int                   `json:"WriteOnly,omitempty"`
	AdditionalProperties map[string]interface{} `json:"-"`
}

// WithoutAdditionalProperties defines model for WithoutAdditionalProperties.
type WithoutAdditionalProperties struct {
	Nullable  *int `json:"Nullable"`
	Optional  *int `json:"Optional,omitempty"`
	ReadOnly  *int `json:"ReadOnly,omitempty"`
	Required  int  `json:"Required"`
	WriteOnly *int `json:"WriteOnly,omitempty"`
}

// Getter for additional properties for WithAdditionalProperties. Returns the specified
// element and whether it was found
func (a WithAdditionalProperties) Get(fieldName string) (value interface{}, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for WithAdditionalProperties
func (a *WithAdditionalProperties) Set(fieldName string, value interface{}) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]interface{})
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for WithAdditionalProperties to handle AdditionalProperties
func (a *WithAdditionalProperties) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if raw, found := object["Nullable"]; found {
		err = json.Unmarshal(raw, &a.Nullable)
		if err != nil {
			return fmt.Errorf("error reading 'Nullable': %w", err)
		}
		delete(object, "Nullable")
	}

	if raw, found := object["Optional"]; found {
		err = json.Unmarshal(raw, &a.Optional)
		if err != nil {
			return fmt.Errorf("error reading 'Optional': %w", err)
		}
		delete(object, "Optional")
	}

	if raw, found := object["ReadOnly"]; found {
		err = json.Unmarshal(raw, &a.ReadOnly)
		if err != nil {
			return fmt.Errorf("error reading 'ReadOnly': %w", err)
		}
		delete(object, "ReadOnly")
	}

	if raw, found := object["Required"]; found {
		err = json.Unmarshal(raw, &a.Required)
		if err != nil {
			return fmt.Errorf("error reading 'Required': %w", err)
		}
		delete(object, "Required")
	}

	if raw, found := object["WriteOnly"]; found {
		err = json.Unmarshal(raw, &a.WriteOnly)
		if err != nil {
			return fmt.Errorf("error reading 'WriteOnly': %w", err)
		}
		delete(object, "WriteOnly")
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]interface{})
		for fieldName, fieldBuf := range object {
			var fieldVal interface{}
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for WithAdditionalProperties to handle AdditionalProperties
func (a WithAdditionalProperties) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	object["Nullable"], err = json.Marshal(a.Nullable)
	if err != nil {
		return nil, fmt.Errorf("error marshaling 'Nullable': %w", err)
	}

	if a.Optional != nil {
		object["Optional"], err = json.Marshal(a.Optional)
		if err != nil {
			return nil, fmt.Errorf("error marshaling 'Optional': %w", err)
		}
	}

	if a.ReadOnly != nil {
		object["ReadOnly"], err = json.Marshal(a.ReadOnly)
		if err != nil {
			return nil, fmt.Errorf("error marshaling 'ReadOnly': %w", err)
		}
	}

	object["Required"], err = json.Marshal(a.Required)
	if err != nil {
		return nil, fmt.Errorf("error marshaling 'Required': %w", err)
	}

	if a.WriteOnly != nil {
		object["WriteOnly"], err = json.Marshal(a.WriteOnly)
		if err != nil {
			return nil, fmt.Errorf("error marshaling 'WriteOnly': %w", err)
		}
	}

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}
