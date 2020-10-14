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
)

func TestStyleParam(t *testing.T) {
	primitive := 5
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
	timestamp, _ := time.Parse(time.RFC3339, "2020-01-01T22:00:00+02:00")

	// ---------------------------- Simple Style -------------------------------

	result, err := StyleParam("simple", false, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "5", result)

	result, err = StyleParam("simple", true, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "5", result)

	result, err = StyleParam("simple", false, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "3,4,5", result)

	result, err = StyleParam("simple", true, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "3,4,5", result)

	result, err = StyleParam("simple", false, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	result, err = StyleParam("simple", true, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex,role=admin", result)

	result, err = StyleParam("simple", false, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	result, err = StyleParam("simple", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex,role=admin", result)

	result, err = StyleParam("simple", false, "id", timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("simple", true, "id", timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("simple", false, "id", &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("simple", true, "id", &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "2020-01-01T22%3A00%3A00%2B02%3A00", result)

	// ----------------------------- Label Style -------------------------------

	result, err = StyleParam("label", false, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".5", result)

	result, err = StyleParam("label", true, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".5", result)

	result, err = StyleParam("label", false, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, ".3,4,5", result)

	result, err = StyleParam("label", true, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, ".3.4.5", result)

	result, err = StyleParam("label", false, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Alex,role,admin", result)

	result, err = StyleParam("label", true, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Alex.role=admin", result)

	result, err = StyleParam("label", false, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Alex,role,admin", result)

	result, err = StyleParam("label", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Alex.role=admin", result)

	result, err = StyleParam("label", false, "id", timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("label", true, "id", timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("label", false, "id", &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("label", true, "id", &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ".2020-01-01T22%3A00%3A00%2B02%3A00", result)

	// ----------------------------- Matrix Style ------------------------------

	result, err = StyleParam("matrix", false, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=5", result)

	result, err = StyleParam("matrix", true, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=5", result)

	result, err = StyleParam("matrix", false, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=3,4,5", result)

	result, err = StyleParam("matrix", true, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=3;id=4;id=5", result)

	result, err = StyleParam("matrix", false, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Alex,role,admin", result)

	result, err = StyleParam("matrix", true, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Alex;role=admin", result)

	result, err = StyleParam("matrix", false, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Alex,role,admin", result)

	result, err = StyleParam("matrix", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Alex;role=admin", result)

	result, err = StyleParam("matrix", false, "id", timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("matrix", true, "id", timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("matrix", false, "id", &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("matrix", true, "id", &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	// ------------------------------ Form Style -------------------------------
	result, err = StyleParam("form", false, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=5", result)

	result, err = StyleParam("form", true, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=5", result)

	result, err = StyleParam("form", false, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3,4,5", result)

	result, err = StyleParam("form", true, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParam("form", false, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Alex,role,admin", result)

	result, err = StyleParam("form", true, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex&role=admin", result)

	result, err = StyleParam("form", false, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Alex,role,admin", result)

	result, err = StyleParam("form", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex&role=admin", result)

	result, err = StyleParam("form", false, "id", timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("form", true, "id", timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("form", false, "id", &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	result, err = StyleParam("form", true, "id", &timestamp)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=2020-01-01T22%3A00%3A00%2B02%3A00", result)

	// ------------------------  spaceDelimited Style --------------------------

	result, err = StyleParam("spaceDelimited", false, "id", primitive)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", true, "id", primitive)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", false, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3 4 5", result)

	result, err = StyleParam("spaceDelimited", true, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParam("spaceDelimited", false, "id", object)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", true, "id", object)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", false, "id", dict)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", true, "id", dict)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", false, "id", timestamp)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", true, "id", timestamp)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", false, "id", &timestamp)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", true, "id", &timestamp)
	assert.Error(t, err)

	// -------------------------  pipeDelimited Style --------------------------

	result, err = StyleParam("pipeDelimited", false, "id", primitive)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", true, "id", primitive)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", false, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3|4|5", result)

	result, err = StyleParam("pipeDelimited", true, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParam("pipeDelimited", false, "id", object)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", true, "id", object)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", false, "id", dict)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", true, "id", dict)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", false, "id", timestamp)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", true, "id", timestamp)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", false, "id", &timestamp)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", true, "id", &timestamp)
	assert.Error(t, err)

	// ---------------------------  deepObject Style ---------------------------
	result, err = StyleParam("deepObject", false, "id", primitive)
	assert.Error(t, err)

	result, err = StyleParam("deepObject", true, "id", primitive)
	assert.Error(t, err)

	result, err = StyleParam("deepObject", false, "id", array)
	assert.Error(t, err)

	result, err = StyleParam("deepObject", true, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[0]=3&id[1]=4&id[2]=5", result)

	result, err = StyleParam("deepObject", false, "id", object)
	assert.Error(t, err)

	result, err = StyleParam("deepObject", true, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Alex&id[role]=admin", result)

	result, err = StyleParam("deepObject", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Alex&id[role]=admin", result)

	result, err = StyleParam("deepObject", false, "id", timestamp)
	assert.Error(t, err)

	result, err = StyleParam("deepObject", true, "id", timestamp)
	assert.Error(t, err)

	result, err = StyleParam("deepObject", false, "id", &timestamp)
	assert.Error(t, err)

	result, err = StyleParam("deepObject", true, "id", &timestamp)
	assert.Error(t, err)

	// Misc tests
	// Test type aliases
	type StrType string
	result, err = StyleParam("simple", false, "foo", StrType("test"))
	assert.NoError(t, err)
	assert.EqualValues(t, "test", result)

	type IntType int32
	result, err = StyleParam("simple", false, "foo", IntType(7))
	assert.NoError(t, err)
	assert.EqualValues(t, "7", result)

	type FloatType float64
	result, err = StyleParam("simple", false, "foo", FloatType(7.5))
	assert.NoError(t, err)
	assert.EqualValues(t, "7.5", result)

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
	result, err = StyleParam("simple", false, "id", object2)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	// Nullable fields need to be excluded when null
	object2.Role = nil
	result, err = StyleParam("simple", false, "id", object2)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex", result)
}
