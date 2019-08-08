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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func escapePassThrough(s string) string {
	return s
}

func TestStyleParam(t *testing.T) {
	primitive := 5
	unescapedPrimitive := "foo=bar"

	array := []int{3, 4, 5}
	unescapedArray := []string{"test?", "Sm!th", "J+mes", "foo=bar"}

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

	unescapedObject := TestObject{
		FirstName: "Al=x",
		Role:      "adm;n",
	}

	// ---------------------------- Simple Style -------------------------------

	result, err := StyleParam("simple", false, escapePassThrough, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "5", result)

	result, err = StyleParam("simple", true, escapePassThrough, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "5", result)

	result, err = StyleParam("simple", false, url.PathEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "foo=bar", result)

	result, err = StyleParam("simple", true, url.PathEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "foo=bar", result)

	result, err = StyleParam("simple", false, url.QueryEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "foo%3Dbar", result)

	result, err = StyleParam("simple", true, url.QueryEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "foo%3Dbar", result)

	result, err = StyleParam("simple", false, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "3,4,5", result)

	result, err = StyleParam("simple", true, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "3,4,5", result)

	result, err = StyleParam("simple", false, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "test%3F,Sm%21th,J+mes,foo=bar", result)

	result, err = StyleParam("simple", true, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "test%3F,Sm%21th,J+mes,foo=bar", result)

	result, err = StyleParam("simple", false, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "test%3F,Sm%21th,J%2Bmes,foo%3Dbar", result)

	result, err = StyleParam("simple", true, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "test%3F,Sm%21th,J%2Bmes,foo%3Dbar", result)

	result, err = StyleParam("simple", false, escapePassThrough, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	result, err = StyleParam("simple", true, escapePassThrough, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex,role=admin", result)

	result, err = StyleParam("simple", false, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	result, err = StyleParam("simple", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex,role=admin", result)
	result, err = StyleParam("simple", false, url.PathEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Al=x,role,adm%3Bn", result)

	result, err = StyleParam("simple", true, url.PathEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Al=x,role=adm%3Bn", result)

	result, err = StyleParam("simple", false, url.QueryEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Al%3Dx,role,adm%3Bn", result)

	result, err = StyleParam("simple", true, url.QueryEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Al%3Dx,role=adm%3Bn", result)

	// ----------------------------- Label Style -------------------------------

	result, err = StyleParam("label", false, escapePassThrough, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".5", result)

	result, err = StyleParam("label", true, escapePassThrough, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".5", result)

	result, err = StyleParam("label", false, url.PathEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".foo=bar", result)

	result, err = StyleParam("label", true, url.PathEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".foo=bar", result)

	result, err = StyleParam("label", false, url.QueryEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".foo%3Dbar", result)

	result, err = StyleParam("label", true, url.QueryEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ".foo%3Dbar", result)

	result, err = StyleParam("label", false, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, ".3,4,5", result)

	result, err = StyleParam("label", true, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, ".3.4.5", result)

	result, err = StyleParam("label", false, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, ".test%3F,Sm%21th,J+mes,foo=bar", result)

	result, err = StyleParam("label", true, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, ".test%3F.Sm%21th.J+mes.foo=bar", result)

	result, err = StyleParam("label", false, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, ".test%3F,Sm%21th,J%2Bmes,foo%3Dbar", result)

	result, err = StyleParam("label", true, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, ".test%3F.Sm%21th.J%2Bmes.foo%3Dbar", result)

	result, err = StyleParam("label", false, escapePassThrough, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Alex,role,admin", result)

	result, err = StyleParam("label", true, escapePassThrough, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Alex.role=admin", result)

	result, err = StyleParam("label", false, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Alex,role,admin", result)

	result, err = StyleParam("label", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Alex.role=admin", result)
	result, err = StyleParam("label", false, url.PathEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Al=x,role,adm%3Bn", result)

	result, err = StyleParam("label", true, url.PathEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Al=x.role=adm%3Bn", result)

	result, err = StyleParam("label", false, url.QueryEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName,Al%3Dx,role,adm%3Bn", result)

	result, err = StyleParam("label", true, url.QueryEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, ".firstName=Al%3Dx.role=adm%3Bn", result)

	// ----------------------------- Matrix Style ------------------------------

	result, err = StyleParam("matrix", false, escapePassThrough, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=5", result)

	result, err = StyleParam("matrix", true, escapePassThrough, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=5", result)

	result, err = StyleParam("matrix", false, url.PathEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=foo=bar", result)

	result, err = StyleParam("matrix", true, url.PathEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=foo=bar", result)

	result, err = StyleParam("matrix", false, url.QueryEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=foo%3Dbar", result)

	result, err = StyleParam("matrix", true, url.QueryEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=foo%3Dbar", result)

	result, err = StyleParam("matrix", false, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=3,4,5", result)

	result, err = StyleParam("matrix", true, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=3;id=4;id=5", result)

	result, err = StyleParam("matrix", false, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=test%3F,Sm%21th,J+mes,foo=bar", result)

	result, err = StyleParam("matrix", true, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=test%3F;id=Sm%21th;id=J+mes;id=foo=bar", result)

	result, err = StyleParam("matrix", false, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=test%3F,Sm%21th,J%2Bmes,foo%3Dbar", result)

	result, err = StyleParam("matrix", true, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=test%3F;id=Sm%21th;id=J%2Bmes;id=foo%3Dbar", result)

	result, err = StyleParam("matrix", false, escapePassThrough, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Alex,role,admin", result)

	result, err = StyleParam("matrix", true, escapePassThrough, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Alex;role=admin", result)

	result, err = StyleParam("matrix", false, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Alex,role,admin", result)

	result, err = StyleParam("matrix", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Alex;role=admin", result)
	result, err = StyleParam("matrix", false, url.PathEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Al=x,role,adm%3Bn", result)

	result, err = StyleParam("matrix", true, url.PathEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Al=x;role=adm%3Bn", result)

	result, err = StyleParam("matrix", false, url.QueryEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, ";id=firstName,Al%3Dx,role,adm%3Bn", result)

	result, err = StyleParam("matrix", true, url.QueryEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, ";firstName=Al%3Dx;role=adm%3Bn", result)

	// ------------------------------ Form Style -------------------------------
	result, err = StyleParam("form", false, escapePassThrough, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=5", result)

	result, err = StyleParam("form", true, escapePassThrough, "id", primitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=5", result)

	result, err = StyleParam("form", false, url.PathEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=foo=bar", result)

	result, err = StyleParam("form", true, url.PathEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=foo=bar", result)

	result, err = StyleParam("form", false, url.QueryEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=foo%3Dbar", result)

	result, err = StyleParam("form", true, url.QueryEscape, "id", unescapedPrimitive)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=foo%3Dbar", result)

	result, err = StyleParam("form", false, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3,4,5", result)

	result, err = StyleParam("form", true, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParam("form", false, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F,Sm%21th,J+mes,foo=bar", result)

	result, err = StyleParam("form", true, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F&id=Sm%21th&id=J+mes&id=foo=bar", result)

	result, err = StyleParam("form", false, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F,Sm%21th,J%2Bmes,foo%3Dbar", result)

	result, err = StyleParam("form", true, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F&id=Sm%21th&id=J%2Bmes&id=foo%3Dbar", result)

	result, err = StyleParam("form", false, escapePassThrough, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Alex,role,admin", result)

	result, err = StyleParam("form", true, escapePassThrough, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex&role=admin", result)

	result, err = StyleParam("form", false, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Alex,role,admin", result)

	result, err = StyleParam("form", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Alex&role=admin", result)
	result, err = StyleParam("form", false, url.PathEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Al=x,role,adm%3Bn", result)

	result, err = StyleParam("form", true, url.PathEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Al=x&role=adm%3Bn", result)

	result, err = StyleParam("form", false, url.QueryEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=firstName,Al%3Dx,role,adm%3Bn", result)

	result, err = StyleParam("form", true, url.QueryEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName=Al%3Dx&role=adm%3Bn", result)

	// ------------------------  spaceDelimited Style --------------------------

	_, err = StyleParam("spaceDelimited", false, escapePassThrough, "id", primitive)
	assert.Error(t, err)

	_, err = StyleParam("spaceDelimited", true, escapePassThrough, "id", primitive)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", false, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3 4 5", result)

	result, err = StyleParam("spaceDelimited", true, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParam("spaceDelimited", false, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F Sm%21th J+mes foo=bar", result)

	result, err = StyleParam("spaceDelimited", true, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F&id=Sm%21th&id=J+mes&id=foo=bar", result)

	result, err = StyleParam("spaceDelimited", false, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F Sm%21th J%2Bmes foo%3Dbar", result)

	result, err = StyleParam("spaceDelimited", true, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F&id=Sm%21th&id=J%2Bmes&id=foo%3Dbar", result)

	_, err = StyleParam("spaceDelimited", false, escapePassThrough, "id", object)
	assert.Error(t, err)

	_, err = StyleParam("spaceDelimited", true, escapePassThrough, "id", object)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", false, "id", dict)
	assert.Error(t, err)

	result, err = StyleParam("spaceDelimited", true, "id", dict)
	assert.Error(t, err)
	// -------------------------  pipeDelimited Style --------------------------

	_, err = StyleParam("pipeDelimited", false, escapePassThrough, "id", primitive)
	assert.Error(t, err)

	_, err = StyleParam("pipeDelimited", true, escapePassThrough, "id", primitive)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", false, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3|4|5", result)

	result, err = StyleParam("pipeDelimited", true, escapePassThrough, "id", array)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=3&id=4&id=5", result)

	result, err = StyleParam("pipeDelimited", false, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F|Sm%21th|J+mes|foo=bar", result)

	result, err = StyleParam("pipeDelimited", true, url.PathEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F&id=Sm%21th&id=J+mes&id=foo=bar", result)

	result, err = StyleParam("pipeDelimited", false, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F|Sm%21th|J%2Bmes|foo%3Dbar", result)

	result, err = StyleParam("pipeDelimited", true, url.QueryEscape, "id", unescapedArray)
	assert.NoError(t, err)
	assert.EqualValues(t, "id=test%3F&id=Sm%21th&id=J%2Bmes&id=foo%3Dbar", result)

	_, err = StyleParam("pipeDelimited", false, escapePassThrough, "id", object)
	assert.Error(t, err)

	_, err = StyleParam("pipeDelimited", true, escapePassThrough, "id", object)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", false, "id", dict)
	assert.Error(t, err)

	result, err = StyleParam("pipeDelimited", true, "id", dict)
	assert.Error(t, err)

	// ---------------------------  deepObject Style ---------------------------
	_, err = StyleParam("deepObject", false, escapePassThrough, "id", primitive)
	assert.Error(t, err)

	_, err = StyleParam("deepObject", true, escapePassThrough, "id", primitive)
	assert.Error(t, err)

	_, err = StyleParam("deepObject", false, escapePassThrough, "id", array)
	assert.Error(t, err)

	_, err = StyleParam("deepObject", true, escapePassThrough, "id", array)
	assert.Error(t, err)

	_, err = StyleParam("deepObject", false, escapePassThrough, "id", object)
	assert.Error(t, err)

	result, err = StyleParam("deepObject", true, escapePassThrough, "id", object)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Alex&id[role]=admin", result)

	result, err = StyleParam("deepObject", true, "id", dict)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Alex&id[role]=admin", result)
	result, err = StyleParam("deepObject", true, url.PathEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Al=x&id[role]=adm%3Bn", result)

	result, err = StyleParam("deepObject", true, url.QueryEscape, "id", unescapedObject)
	assert.NoError(t, err)
	assert.EqualValues(t, "id[firstName]=Al%3Dx&id[role]=adm%3Bn", result)

	// Misc tests
	// Test type aliases
	type StrType string
	result, err = StyleParam("simple", false, escapePassThrough, "foo", StrType("test"))
	assert.NoError(t, err)
	assert.EqualValues(t, "test", result)

	type IntType int32
	result, err = StyleParam("simple", false, escapePassThrough, "foo", IntType(7))
	assert.NoError(t, err)
	assert.EqualValues(t, "7", result)

	type FloatType float64
	result, err = StyleParam("simple", false, escapePassThrough, "foo", FloatType(7.5))
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
	result, err = StyleParam("simple", false, escapePassThrough, "id", object2)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex,role,admin", result)

	// Nullable fields need to be excluded when null
	object2.Role = nil
	result, err = StyleParam("simple", false, escapePassThrough, "id", object2)
	assert.NoError(t, err)
	assert.EqualValues(t, "firstName,Alex", result)
}
