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

	"github.com/stretchr/testify/assert"
)

func TestSplitParameter(t *testing.T) {
	// Please see the parameter serialization docs to understand these test
	// cases

	expectedPrimitive := []string{"5"}
	expectedArray := []string{"3", "4", "5"}
	expectedObject := []string{"role", "admin", "firstName", "Alex"}
	expectedExplodedObject := []string{"role=admin", "firstName=Alex"}

	var result []string
	var err error
	//  ------------------------ simple style ---------------------------------
	result, err = splitStyledParameter("simple",
		false,
		false,
		"id",
		"5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedPrimitive, result)

	result, err = splitStyledParameter("simple",
		false,
		false,
		"id",
		"3,4,5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedArray, result)

	result, err = splitStyledParameter("simple",
		false,
		true,
		"id",
		"role,admin,firstName,Alex")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedObject, result)

	result, err = splitStyledParameter("simple",
		true,
		false,
		"id",
		"5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedPrimitive, result)

	result, err = splitStyledParameter("simple",
		true,
		false,
		"id",
		"3,4,5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedArray, result)

	result, err = splitStyledParameter("simple",
		true,
		true,
		"id",
		"role=admin,firstName=Alex")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedExplodedObject, result)

	//  ------------------------ label style ---------------------------------
	result, err = splitStyledParameter("label",
		false,
		false,
		"id",
		".5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedPrimitive, result)

	result, err = splitStyledParameter("label",
		false,
		false,
		"id",
		".3,4,5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedArray, result)

	result, err = splitStyledParameter("label",
		false,
		true,
		"id",
		".role,admin,firstName,Alex")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedObject, result)

	result, err = splitStyledParameter("label",
		true,
		false,
		"id",
		".5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedPrimitive, result)

	result, err = splitStyledParameter("label",
		true,
		false,
		"id",
		".3.4.5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedArray, result)

	result, err = splitStyledParameter("label",
		true,
		true,
		"id",
		".role=admin.firstName=Alex")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedExplodedObject, result)

	//  ------------------------ matrix style ---------------------------------
	result, err = splitStyledParameter("matrix",
		false,
		false,
		"id",
		";id=5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedPrimitive, result)

	result, err = splitStyledParameter("matrix",
		false,
		false,
		"id",
		";id=3,4,5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedArray, result)

	result, err = splitStyledParameter("matrix",
		false,
		true,
		"id",
		";id=role,admin,firstName,Alex")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedObject, result)

	result, err = splitStyledParameter("matrix",
		true,
		false,
		"id",
		";id=5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedPrimitive, result)

	result, err = splitStyledParameter("matrix",
		true,
		false,
		"id",
		";id=3;id=4;id=5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedArray, result)

	result, err = splitStyledParameter("matrix",
		true,
		true,
		"id",
		";role=admin;firstName=Alex")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedExplodedObject, result)

	// ------------------------ form style ---------------------------------
	result, err = splitStyledParameter("form",
		false,
		false,
		"id",
		"id=5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedPrimitive, result)

	result, err = splitStyledParameter("form",
		false,
		false,
		"id",
		"id=3,4,5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedArray, result)

	result, err = splitStyledParameter("form",
		false,
		true,
		"id",
		"id=role,admin,firstName,Alex")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedObject, result)

	result, err = splitStyledParameter("form",
		true,
		false,
		"id",
		"id=5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedPrimitive, result)

	result, err = splitStyledParameter("form",
		true,
		false,
		"id",
		"id=3&id=4&id=5")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedArray, result)

	result, err = splitStyledParameter("form",
		true,
		true,
		"id",
		"role=admin&firstName=Alex")
	assert.NoError(t, err)
	assert.EqualValues(t, expectedExplodedObject, result)
}
