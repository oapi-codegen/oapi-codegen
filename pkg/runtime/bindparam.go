// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
)

// This function binds a parameter as described in the Path Parameters
// section here to a Go object:
// https://swagger.io/docs/specification/serialization/
func BindStyledParameter(style string, explode bool, paramName string,
	value string, dest interface{}) error {

	if value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "parameter '%s' is empty, can't bind its value", paramName)
	}

	// Everything comes in by pointer, dereference it
	v := reflect.Indirect(reflect.ValueOf(dest))

	// This is the basic type of the destination object.
	t := v.Type()

	if t.Kind() == reflect.Struct {
		// We've got a destination object, we'll create a JSON representation
		// of the input value, and let the json library deal with the unmarshaling
		parts, err := splitStyledParameter(style, explode, true, paramName, value)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = bindSplitPartsToDestinationStruct(paramName, parts, explode, dest)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return nil
	}

	if t.Kind() == reflect.Slice {
		// Chop up the parameter into parts based on its style
		parts, err := splitStyledParameter(style, explode, false, paramName, value)
		if err != nil {
			return fmt.Errorf("error splitting input '%s' into parts: %s", value, err)
		}

		err = bindSplitPartsToDestinationArray(parts, dest)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return nil
	}

	// Try to bind the remaining types as a base type.
	return BindStringToObject(value, dest)
}

// This is a complex set of operations, but each given parameter style can be
// packed together in multiple ways, using different styles of separators, and
// different packing strategies based on the explode flag. This function takes
// as input any parameter format, and unpacks it to a simple list of strings
// or key-values which we can then treat generically.
// Why, oh why, great Swagger gods, did you have to make this so complicated?
func splitStyledParameter(style string, explode bool, object bool, paramName string, value string) ([]string, error) {
	switch style {
	case "simple":
		// In the simple case, we always split on comma
		parts := strings.Split(value, ",")
		return parts, nil
	case "label":
		// In the label case, it's more tricky. In the no explode case, we have
		// /users/.3,4,5 for arrays
		// /users/.role,admin,firstName,Alex for objects
		// in the explode case, we have:
		// /users/.3.4.5
		// /users/.role=admin.firstName=Alex
		if explode {
			// In the exploded case, split everything on periods.
			parts := strings.Split(value, ".")
			// The first part should be an empty string because we have a
			// leading period.
			if parts[0] != "" {
				return nil, fmt.Errorf("invalid format for label parameter '%s', should start with '.'", paramName)
			}
			return parts[1:], nil

		} else {
			// In the unexploded case, we strip off the leading period.
			if value[0] != '.' {
				return nil, fmt.Errorf("invalid format for label parameter '%s', should start with '.'", paramName)
			}
			// The rest is comma separated.
			return strings.Split(value[1:], ","), nil
		}

	case "matrix":
		if explode {
			// In the exploded case, we break everything up on semicolon
			parts := strings.Split(value, ";")
			// The first part should always be empty string, since we started
			// with ;something
			if parts[0] != "" {
				return nil, fmt.Errorf("invalid format for matrix parameter '%s', should start with ';'", paramName)
			}
			parts = parts[1:]
			// Now, if we have an object, we just have a list of x=y statements.
			// for a non-object, like an array, we have id=x, id=y. id=z, etc,
			// so we need to strip the prefix from each of them.
			if !object {
				prefix := paramName + "="
				for i := range parts {
					parts[i] = strings.TrimPrefix(parts[i], prefix)
				}
			}
			return parts, nil
		} else {
			// In the unexploded case, parameters will start with ;paramName=
			prefix := ";" + paramName + "="
			if !strings.HasPrefix(value, prefix) {
				return nil, fmt.Errorf("expected parameter '%s' to start with %s", paramName, prefix)
			}
			str := strings.TrimPrefix(value, prefix)
			return strings.Split(str, ","), nil
		}
	case "form":
		var parts []string
		if explode {
			parts = strings.Split(value, "&")
			if !object {
				prefix := paramName + "="
				for i := range parts {
					parts[i] = strings.TrimPrefix(parts[i], prefix)
				}
			}
			return parts, nil
		} else {
			parts = strings.Split(value, ",")
			prefix := paramName + "="
			for i := range parts {
				parts[i] = strings.TrimPrefix(parts[i], prefix)
			}
		}
		return parts, nil
	}

	return nil, fmt.Errorf("unhandled parameter style: %s", style)
}

// Given a set of values as a slice, create a slice to hold them all, and
// assign to each one by one.
func bindSplitPartsToDestinationArray(parts []string, dest interface{}) error {
	// Everything comes in by pointer, dereference it
	v := reflect.Indirect(reflect.ValueOf(dest))

	// This is the basic type of the destination object.
	t := v.Type()

	// We've got a destination array, bind each object one by one.
	// This generates a slice of the correct element type and length to
	// hold all the parts.
	newArray := reflect.MakeSlice(t, len(parts), len(parts))
	for i, p := range parts {
		err := BindStringToObject(p, newArray.Index(i).Addr().Interface())
		if err != nil {
			return fmt.Errorf("error setting array element: %s", err)
		}
	}
	v.Set(newArray)
	return nil
}

// Given a set of chopped up parameter parts, bind them to a destination
// struct. The exploded parameter controls whether we send key value pairs
// in the exploded case, or a sequence of values which are interpreted as
// tuples.
// Given the struct Id { firstName string, role string }, as in the canonical
// swagger examples, in the exploded case, we would pass
// ["firstName=Alex", "role=admin"], where in the non-exploded case, we would
// pass "firstName", "Alex", "role", "admin"]
//
// We punt the hard work of binding these values to the object to the json
// library. We'll turn those arrays into JSON strings, and unmarshal
// into the struct.
func bindSplitPartsToDestinationStruct(paramName string, parts []string, explode bool, dest interface{}) error {
	// We've got a destination object, we'll create a JSON representation
	// of the input value, and let the json library deal with the unmarshaling
	var fields []string
	if explode {
		fields = make([]string, len(parts))
		for i, property := range parts {
			propertyParts := strings.Split(property, "=")
			if len(propertyParts) != 2 {
				return fmt.Errorf("parameter '%s' has invalid exploded format", paramName)
			}
			fields[i] = "\"" + propertyParts[0] + "\":\"" + propertyParts[1] + "\""
		}
	} else {
		if len(parts)%2 != 0 {
			return fmt.Errorf("parameter '%s' has invalid format, property/values need to be pairs", paramName)
		}
		fields = make([]string, len(parts)/2)
		for i := 0; i < len(parts); i += 2 {
			key := parts[i]
			value := parts[i+1]
			fields[i/2] = "\"" + key + "\":\"" + value + "\""
		}
	}
	jsonParam := "{" + strings.Join(fields, ",") + "}"
	err := json.Unmarshal([]byte(jsonParam), dest)
	if err != nil {
		return fmt.Errorf("error binding parameter %s fields: %s", paramName, err)
	}
	return nil
}

// This works much like BindStyledParameter, however it takes a query argument
// input array from the url package, since query arguments come through a
// different path than the styled arguments. They're also exceptionally fussy.
// For example, consider the exploded and unexploded form parameter examples:
// (exploded) /users?role=admin&firstName=Alex
// (unexploded) /users?id=role,admin,firstName,Alex
//
// In the first case, we can pull the "id" parameter off the echo context,
// and unmarshal via json as an intermediate. Easy. In the second case, we
// don't have the id QueryParam present, but must find "role", and "firstName".
// what if there is another parameter similar to "ID" named "role"? We can't
// tell them apart. This code tries to fail, but the moral of the story is that
// you shouldn't pass objects via form styled query arguments, just use
// the Content parameter form.
func BindQueryParameter(style string, explode bool, required bool, paramName string,
	queryParams url.Values, dest interface{}) error {

	// dv = destination value.
	dv := reflect.Indirect(reflect.ValueOf(dest))

	// intermediate value form which is either dv or dv dereferenced.
	v := dv

	// inner code will bind the string's value to this interface.
	var output interface{}

	if required {
		// If the parameter is required, then the generated code will pass us
		// a pointer to it: &int, &object, and so forth. We can directly set
		// them.
		output = dest
	} else {
		// For optional parameters, we have an extra indirect. An optional
		// parameter of type "int" will be *int on the struct. We pass that
		// in by pointer, and have **int.

		// If the destination, is a nil pointer, we need to allocate it.
		if v.IsNil() {
			t := v.Type()
			newValue := reflect.New(t.Elem())
			// for now, hang onto the output buffer separately from destination,
			// as we don't want to write anything to destination until we can
			// unmarshal successfully, and check whether a field is required.
			output = newValue.Interface()
		} else {
			// If the destination isn't nil, just use that.
			output = v.Interface()
		}

		// Get rid of that extra indirect as compared to the required case,
		// so the code below doesn't have to care.
		v = reflect.Indirect(reflect.ValueOf(output))
	}

	// This is the basic type of the destination object.
	t := v.Type()
	k := t.Kind()

	switch style {
	case "form":
		var parts []string
		if explode {
			// ok, the explode case in query arguments is very, very annoying,
			// because an exploded object, such as /users?role=admin&firstName=Alex
			// isn't actually present in the parameter array. We have to do
			// different things based on destination type.
			values, found := queryParams[paramName]
			var err error
			switch k {
			case reflect.Slice:
				// In the slice case, we simply use the arguments provided by
				// http library.
				if !found {
					if required {
						return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
							"query parameter '%s' is required", paramName))
					} else {
						return nil
					}
				}
				err = bindSplitPartsToDestinationArray(values, output)
			case reflect.Struct:
				// This case is really annoying, and error prone, but the
				// form style object binding doesn't tell us which arguments
				// in the query string correspond to the object's fields. We'll
				// try to bind field by field.
				err = bindParamsToExplodedObject(paramName, queryParams, output)
			default:
				// Primitive object case. We expect to have 1 value to
				// unmarshal.
				if len(values) == 0 {
					if required {
						return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
							"query parameter '%s' is required", paramName))
					} else {
						return nil
					}
				}
				if len(values) != 1 {
					return echo.NewHTTPError(http.StatusBadRequest,
						fmt.Sprintf("multiple values for single value parameter '%s'", paramName))
				}
				err = BindStringToObject(values[0], output)
			}
			if err != nil {
				return err
			}
			// If the parameter is required, and we've successfully unmarshaled
			// it, this assigns the new object to the pointer pointer.
			if !required {
				dv.Set(reflect.ValueOf(output))
			}
			return nil
		} else {
			values, found := queryParams[paramName]
			if !found {
				if required {
					return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
						"query parameter '%s' is required", paramName))
				} else {
					return nil
				}
			}
			if len(values) != 1 {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("parameter '%s' is not exploded, but is specified multiple times", paramName))
			}
			parts = strings.Split(values[0], ",")
		}
		var err error
		switch k {
		case reflect.Slice:
			err = bindSplitPartsToDestinationArray(parts, output)
		case reflect.Struct:
			err = bindSplitPartsToDestinationStruct(paramName, parts, explode, output)
		default:
			if len(parts) == 0 {
				if required {
					return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
						"query parameter '%s' is required", paramName))
				} else {
					return nil
				}
			}
			if len(parts) != 1 {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("multiple values for single value parameter '%s'", paramName))
			}
			err = BindStringToObject(parts[0], output)
		}
		if err != nil {
			return err
		}
		if !required {
			dv.Set(reflect.ValueOf(output))
		}
		return nil
	case "deepObject":
		// Loop through all queryParams and fill the objectMap with all key/value pairs having paramName as key
		objectMap := map[string]string{}
		for k, v := range queryParams {
			if !strings.HasPrefix(k, paramName + "[") {
				continue
			}
			split := strings.Split(k, "[")
			if len(split) != 2 {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("parameter '%s=%s' does not match deepObject style", k, v))
			}

			k = strings.TrimSuffix(split[1], "]")
			objectMap[k] = v[0]
		}

		// Marshal and unmarshal the objectMap into dest
		data, err := json.Marshal(objectMap)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(data, dest); err != nil {
			return err
		}
		return nil
	case "spaceDelimited", "pipeDelimited":
		return echo.NewHTTPError(http.StatusNotImplemented,
			fmt.Sprintf("query arguments of style '%s' aren't yet supported", style))
	default:
		return echo.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("style '%s' on parameter '%s' is invalid", style, paramName))

	}
}

// This function reflects the destination structure, and pulls the value for
// each settable field from the given parameters map. This is to deal with the
// exploded form styled object which may occupy any number of parameter names.
// We don't try to be smart here, if the field exists as a query argument,
// set its value.
func bindParamsToExplodedObject(paramName string, values url.Values, dest interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(dest))
	if v.Type().Kind() != reflect.Struct {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"unmarshaling query arg '%s' into wrong type", paramName)
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldT := t.Field(i)

		// Skip unsettable fields, such as internal ones.
		if !v.Field(i).CanSet() {
			continue
		}

		// Find the json annotation on the field, and use the json specified
		// name if available, otherwise, just the field name.
		tag := fieldT.Tag.Get("json")
		fieldName := fieldT.Name
		if tag != "" {
			tagParts := strings.Split(tag, ",")
			name := tagParts[0]
			if name != "" {
				fieldName = name
			}
		}

		// At this point, we look up field name in the parameter list.
		fieldVal, found := values[fieldName]
		if found {
			if len(fieldVal) != 1 {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("field '%s' specified multiple times for param '%s'", fieldName, paramName))
			}
			err := BindStringToObject(fieldVal[0], v.Field(i).Addr().Interface())
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("could not bind query arg '%s' to request object: %s", paramName, err))
			}
		}
	}
	return nil
}
