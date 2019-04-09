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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
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

	// Time won't work with the generic switch below, so handle it separately.
	if dstTime, ok := dst.(*time.Time); ok {
		parsedTime, err := time.Parse(time.RFC3339Nano, src)
		if err != nil {
			return fmt.Errorf("error parsing '%s' as RFC3339 time: %s", src, err)
		}
		*dstTime = parsedTime
		return nil
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
	default:
		// We've got a bunch of types unimplemented, don't fail silently.
		err = fmt.Errorf("can not bind to destination of type: %s", t.Kind())
	}
	if err != nil {
		return fmt.Errorf("error binding string parameter: %s", err)
	}
	return nil
}
