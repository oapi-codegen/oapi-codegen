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

	"github.com/stretchr/testify/assert"

	"github.com/deepmap/oapi-codegen/pkg/types"
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

	result, err := StyleParamWithLocation("simple", false, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "5", result)

	result, err = StyleParamWithLocation("simple", true, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "5", result)

	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "3,4,5", result)

	result, err = StyleParamWithLocation("simple", true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "3,4,5", result)

	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation("simple", true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex,role=admin", result)

	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation("simple", true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex,role=admin", result)

	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("simple", true, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("simple", true, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01", result)

	result, err = StyleParamWithLocation("simple", true, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01", result)

	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01", result)

	result, err = StyleParamWithLocation("simple", true, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01", result)

	// ----------------------------- Label Style -------------------------------

	result, err = StyleParamWithLocation("label", false, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".5", result)

	result, err = StyleParamWithLocation("label", true, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".5", result)

	result, err = StyleParamWithLocation("label", false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, ".3,4,5", result)

	result, err = StyleParamWithLocation("label", true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, ".3.4.5", result)

	result, err = StyleParamWithLocation("label", false, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation("label", true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Alex.role=admin", result)

	result, err = StyleParamWithLocation("label", false, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation("label", true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Alex.role=admin", result)

	result, err = StyleParamWithLocation("label", false, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("label", true, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("label", false, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("label", true, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("label", false, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01", result)

	result, err = StyleParamWithLocation("label", true, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01", result)

	result, err = StyleParamWithLocation("label", false, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01", result)

	result, err = StyleParamWithLocation("label", true, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01", result)

	// ----------------------------- Matrix Style ------------------------------

	result, err = StyleParamWithLocation("matrix", false, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=5", result)

	result, err = StyleParamWithLocation("matrix", true, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=5", result)

	result, err = StyleParamWithLocation("matrix", false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=3,4,5", result)

	result, err = StyleParamWithLocation("matrix", true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=3;id=4;id=5", result)

	result, err = StyleParamWithLocation("matrix", false, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation("matrix", true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Alex;role=admin", result)

	result, err = StyleParamWithLocation("matrix", false, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation("matrix", true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Alex;role=admin", result)

	result, err = StyleParamWithLocation("matrix", false, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("matrix", true, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("matrix", false, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("matrix", true, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("matrix", false, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01", result)

	result, err = StyleParamWithLocation("matrix", true, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01", result)

	result, err = StyleParamWithLocation("matrix", false, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01", result)

	result, err = StyleParamWithLocation("matrix", true, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01", result)

	// ------------------------------ Form Style -------------------------------
	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=5", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=5", result)

	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, primitiveString)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=123", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, primitiveString)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=123", result)

	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, primitiveStringWithReservedChar)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=123%3B456", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, primitiveStringWithReservedChar)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=123%3B456", result)

	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3,4,5", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex&role=admin", result)

	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Alex,role,admin", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex&role=admin", result)

	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, date)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01", result)

	result, err = StyleParamWithLocation("form", false, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01", result)

	result, err = StyleParamWithLocation("form", true, "id", ParamLocationQuery, &date)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01", result)

	// ------------------------  spaceDelimited Style --------------------------

	result, err = StyleParamWithLocation("spaceDelimited", false, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", true, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3 4 5", result)

	result, err = StyleParamWithLocation("spaceDelimited", true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParamWithLocation("spaceDelimited", false, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", true, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", false, "id", ParamLocationQuery, dict)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", true, "id", ParamLocationQuery, dict)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", false, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", true, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", false, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", true, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", false, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", true, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", false, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("spaceDelimited", true, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	// -------------------------  pipeDelimited Style --------------------------

	result, err = StyleParamWithLocation("pipeDelimited", false, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", true, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", false, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3|4|5", result)

	result, err = StyleParamWithLocation("pipeDelimited", true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParamWithLocation("pipeDelimited", false, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", true, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", false, "id", ParamLocationQuery, dict)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", true, "id", ParamLocationQuery, dict)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", false, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", true, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", false, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", true, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", false, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", true, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", false, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("pipeDelimited", true, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	// ---------------------------  deepObject Style ---------------------------
	result, err = StyleParamWithLocation("deepObject", false, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", true, "id", ParamLocationQuery, primitive)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", false, "id", ParamLocationQuery, array)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", true, "id", ParamLocationQuery, array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[0]=3&id[1]=4&id[2]=5", result)

	result, err = StyleParamWithLocation("deepObject", false, "id", ParamLocationQuery, object)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", true, "id", ParamLocationQuery, object)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Alex&id[role]=admin", result)

	result, err = StyleParamWithLocation("deepObject", true, "id", ParamLocationQuery, dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Alex&id[role]=admin", result)

	result, err = StyleParamWithLocation("deepObject", false, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", true, "id", ParamLocationQuery, timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", false, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", true, "id", ParamLocationQuery, &timestamp)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", false, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", true, "id", ParamLocationQuery, date)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", false, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	result, err = StyleParamWithLocation("deepObject", true, "id", ParamLocationQuery, &date)
	assert.Error(t, err)

	// Misc tests
	// Test type aliases
	type StrType string
	result, err = StyleParamWithLocation("simple", false, "foo", ParamLocationQuery, StrType("test"))
	assert.NoError(t, err)
	assert.EqualValues(t, "test", result)

	type IntType int32
	result, err = StyleParamWithLocation("simple", false, "foo", ParamLocationQuery, IntType(7))
	assert.NoError(t, err)
	assert.EqualValues(t, "7", result)

	type UintType uint
	result, err = StyleParamWithLocation("simple", false, "foo", ParamLocationQuery, UintType(9))
	assert.NoError(t, err)
	assert.EqualValues(t, "9", result)

	type Uint8Type uint8
	result, err = StyleParamWithLocation("simple", false, "foo", ParamLocationQuery, Uint8Type(9))
	assert.NoError(t, err)
	assert.EqualValues(t, "9", result)

	type Uint16Type uint16
	result, err = StyleParamWithLocation("simple", false, "foo", ParamLocationQuery, Uint16Type(9))
	assert.NoError(t, err)
	assert.EqualValues(t, "9", result)

	type Uint32Type uint32
	result, err = StyleParamWithLocation("simple", false, "foo", ParamLocationQuery, Uint32Type(9))
	assert.NoError(t, err)
	assert.EqualValues(t, "9", result)

	type Uint64Type uint64
	result, err = StyleParamWithLocation("simple", false, "foo", ParamLocationQuery, Uint64Type(9))
	assert.NoError(t, err)
	assert.EqualValues(t, "9", result)

	type FloatType64 float64
	result, err = StyleParamWithLocation("simple", false, "foo", ParamLocationQuery, FloatType64(7.5))
	assert.NoError(t, err)
	assert.EqualValues(t, "7.5", result)

	type FloatType32 float32
	result, err = StyleParamWithLocation("simple", false, "foo", ParamLocationQuery, FloatType32(1.05))
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
	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, object2)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	// Nullable fields need to be excluded when null
	object2.Role = nil
	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, object2)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex", result)

	// Test handling of time and date inside objects
	type testObject3 struct {
		TimeField time.Time  `json:"time_field"`
		DateField types.Date `json:"date_field"`
	}
	timeVal := time.Date(1996, time.March, 19, 0, 0, 0, 0, time.UTC)
	dateVal := types.Date{
		Time: timeVal,
	}

	object3 := testObject3{
		TimeField: timeVal,
		DateField: dateVal,
	}

	result, err = StyleParamWithLocation("simple", false, "id", ParamLocationQuery, object3)
	assert.NoError(t, err)
	assert.EqualValues(t, "date_field,1996-03-19,time_field,1996-03-19T00%3A00%3A00Z", result)
}
