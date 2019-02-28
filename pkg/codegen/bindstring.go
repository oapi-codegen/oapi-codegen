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
package codegen

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// This function takes a string, and attempts to assign it to the destination
// interface via whatever type conversion is necessary. We have to do this
// via reflection instead of a much simpler type switch so that we can handle
// type aliases. This function was the easy way out, the better way, since we
// know the destination type each place that we use this, is to generate code
// to read each specific type.
func BindStringToObject(src string, dst interface{}) error {
	var err error

	v := reflect.ValueOf(dst)
	t := reflect.TypeOf(dst)

	// We need to dereference pointers
	if t.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		t = v.Type()
	}

	// The resulting type must be settable. reflect will catch issues like
	// passing the destination by value.
	if !v.CanSet() {
		return errors.New("destination is not settable")
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		var val int64
		val, err = strconv.ParseInt(src, 10, 64)
		if err == nil {
			v.SetInt(val)
		}
	case reflect.String:
		v.SetString(src)
		err = nil
	case reflect.Float64, reflect.Float32:
		var val float64
		val, err = strconv.ParseFloat(src, 64)
		if err == nil {
			v.SetFloat(val)
		}
	case reflect.Bool:
		var val bool
		val, err = strconv.ParseBool(src)
		if err == nil {
			v.SetBool(val)
		}
	case reflect.Struct:
		err = json.Unmarshal([]byte(src), dst)
	case reflect.Slice:
		// For slices, we assume a form-type input array, meaning a comma
		// separated string. We split the string on comma to get a bunch of
		// string parts.
		// TODO: Parameterize the format, since Swagger supports several
		//       parameter array formats.
		parts := strings.Split(src, ",")
		// This generates a slice of the correct element type and length to
		// hold all the parts.
		newArray := reflect.MakeSlice(t, len(parts), len(parts))
		// For all the parts, just call ourselves recursively, binding each
		// individual element from its source string.
		for i, p := range parts {
			err = BindStringToObject(p, newArray.Index(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("error setting array element: %s", err)
			}
		}
		v.Set(newArray)
	default:
		// We've got a bunch of types unimplemented, don't fail silently.
		err = fmt.Errorf("can not bind to destination of type: %s", t.Kind())
	}
	if err != nil {
		return fmt.Errorf("error binding string parameter: %s", err)
	}
	return nil
}
