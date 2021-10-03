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
	"testing"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestStyleParam(t *testing.T) {
	primitive := 5
	primitiveString := "123"
	primitiveStringWithReservedChar := "123;456"
	array := []int{3, 4, 5}
	type TestObject struct {
		FirstName string `json:"firstName"`
		Role      string `json:"role"`
	}
	object := TestObject{
		FirstName: "Alex",
		Role:      "admin",
	}
	dict := map[string]interface{}{}
	dict["firstName"] = "Alex"
	dict["role"] = "admin"

	type AliasedTime time.Time
	ti, _ := time.Parse(time.RFC3339, "2020-01-01T22:00:00+02:00")
	timestamp := AliasedTime(ti)

	type AliasedDate types.Date
	date := AliasedDate{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}

	// ---------------------------- Simple Style -------------------------------

	result, err := StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "5", result)

	result, err = StyleParamWithLocation(styleSimple, true, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "5", result)

	result, err = StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "3,4,5", result)

	result, err = StyleParamWithLocation(styleSimple, true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "3,4,5", result)

	result, err = StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation(styleSimple, true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex,role=admin", result)

	result, err = StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation(styleSimple, true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex,role=admin", result)

	result, err = StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleSimple, true, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleSimple, true, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01", result)

	result, err = StyleParamWithLocation(styleSimple, true, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01", result)

	result, err = StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01", result)

	result, err = StyleParamWithLocation(styleSimple, true, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01", result)

	// ----------------------------- Label Style -------------------------------

	result, err = StyleParamWithLocation(styleLabel, false, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".5", result)

	result, err = StyleParamWithLocation(styleLabel, true, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".5", result)

	result, err = StyleParamWithLocation(styleLabel, false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, ".3,4,5", result)

	result, err = StyleParamWithLocation(styleLabel, true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, ".3.4.5", result)

	result, err = StyleParamWithLocation(styleLabel, false, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation(styleLabel, true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Alex.role=admin", result)

	result, err = StyleParamWithLocation(styleLabel, false, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation(styleLabel, true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Alex.role=admin", result)

	result, err = StyleParamWithLocation(styleLabel, false, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleLabel, true, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleLabel, false, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleLabel, true, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleLabel, false, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01", result)

	result, err = StyleParamWithLocation(styleLabel, true, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01", result)

	result, err = StyleParamWithLocation(styleLabel, false, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01", result)

	result, err = StyleParamWithLocation(styleLabel, true, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01", result)

	// ----------------------------- Matrix Style ------------------------------

	result, err = StyleParamWithLocation(styleMatrix, false, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=5", result)

	result, err = StyleParamWithLocation(styleMatrix, true, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=5", result)

	result, err = StyleParamWithLocation(styleMatrix, false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=3,4,5", result)

	result, err = StyleParamWithLocation(styleMatrix, true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=3;id=4;id=5", result)

	result, err = StyleParamWithLocation(styleMatrix, false, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation(styleMatrix, true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Alex;role=admin", result)

	result, err = StyleParamWithLocation(styleMatrix, false, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation(styleMatrix, true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Alex;role=admin", result)

	result, err = StyleParamWithLocation(styleMatrix, false, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleMatrix, true, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleMatrix, false, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleMatrix, true, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleMatrix, false, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01", result)

	result, err = StyleParamWithLocation(styleMatrix, true, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01", result)

	result, err = StyleParamWithLocation(styleMatrix, false, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01", result)

	result, err = StyleParamWithLocation(styleMatrix, true, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01", result)

	// ------------------------------ Form Style -------------------------------
	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=5", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=5", result)

	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, primitiveString)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=123", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, primitiveString)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=123", result)

	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, primitiveStringWithReservedChar)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=123%3B456", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, primitiveStringWithReservedChar)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=123%3B456", result)

	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3,4,5", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex&role=admin", result)

	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex&role=admin", result)

	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01", result)

	result, err = StyleParamWithLocation(styleForm, false, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01", result)

	result, err = StyleParamWithLocation(styleForm, true, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01", result)

	// ------------------------  spaceDelimited Style --------------------------

	_, err = StyleParamWithLocation(styleSpaceDelimited, false, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, true, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	result, err = StyleParamWithLocation(styleSpaceDelimited, false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3 4 5", result)

	result, err = StyleParamWithLocation(styleSpaceDelimited, true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	_, err = StyleParamWithLocation(styleSpaceDelimited, false, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, true, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, false, "id", ParamLocationQuery, dict)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, true, "id", ParamLocationQuery, dict)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, false, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, true, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, false, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, true, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, false, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, true, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, false, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleSpaceDelimited, true, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	// -------------------------  pipeDelimited Style --------------------------

	_, err = StyleParamWithLocation(stylePipeDelimited, false, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, true, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	result, err = StyleParamWithLocation(stylePipeDelimited, false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3|4|5", result)

	result, err = StyleParamWithLocation(stylePipeDelimited, true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	_, err = StyleParamWithLocation(stylePipeDelimited, false, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, true, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, false, "id", ParamLocationQuery, dict)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, true, "id", ParamLocationQuery, dict)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, false, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, true, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, false, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, true, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, false, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, true, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, false, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(stylePipeDelimited, true, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	// ---------------------------  deepObject Style ---------------------------
	_, err = StyleParamWithLocation(styleDeepObject, false, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleDeepObject, true, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleDeepObject, false, "id", ParamLocationQuery, array)
	assert.Error(t, err)

	result, err = StyleParamWithLocation(styleDeepObject, true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[0]=3&id[1]=4&id[2]=5", result)

	_, err = StyleParamWithLocation(styleDeepObject, false, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	result, err = StyleParamWithLocation(styleDeepObject, true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Alex&id[role]=admin", result)

	result, err = StyleParamWithLocation(styleDeepObject, true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Alex&id[role]=admin", result)

	_, err = StyleParamWithLocation(styleDeepObject, false, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleDeepObject, true, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleDeepObject, false, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleDeepObject, true, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleDeepObject, false, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleDeepObject, true, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleDeepObject, false, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	_, err = StyleParamWithLocation(styleDeepObject, true, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	// Misc tests
	// Test type aliases
	type StrType string
	result, err = StyleParamWithLocation(styleSimple, false, "foo", ParamLocationQuery, StrType("test"))
	assert.NoError(t, err)
	assert.EqualValues(t, "test", result)

	type IntType int32
	result, err = StyleParamWithLocation(styleSimple, false, "foo", ParamLocationQuery, IntType(7))
	assert.NoError(t, err)
	assert.EqualValues(t, "7", result)

	type FloatType64 float64
	result, err = StyleParamWithLocation(styleSimple, false, "foo", ParamLocationQuery, FloatType64(7.5))
	assert.NoError(t, err)
	assert.EqualValues(t, "7.5", result)

	type FloatType32 float32
	result, err = StyleParamWithLocation(styleSimple, false, "foo", ParamLocationQuery, FloatType32(1.05))
	assert.NoError(t, err)
	assert.EqualValues(t, "1.05", result)

	// Test that we handle optional fields
	type TestObject2 struct {
		FirstName *string `json:"firstName"`
		Role      *string `json:"role"`
	}
	name := "Alex"
	role := "admin"
	object2 := TestObject2{
		FirstName: &name,
		Role:      &role,
	}
	result, err = StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, object2)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	// Nullable fields need to be excluded when null
	object2.Role = nil
	result, err = StyleParamWithLocation(styleSimple, false, "id", ParamLocationQuery, object2)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex", result)
}
