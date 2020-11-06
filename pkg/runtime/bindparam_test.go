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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deepmap/oapi-codegen/pkg/types"
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

func TestBindQueryParameter(t *testing.T) {
	t.Run("deepObject", func(t *testing.T) {
		type ID struct {
			FirstName *string     `json:"firstName"`
			LastName  *string     `json:"lastName"`
			Role      string      `json:"role"`
			Birthday  *types.Date `json:"birthday"`
		}

		expectedName := "Alex"
		expectedDeepObject := &ID{
			FirstName: &expectedName,
			Role:      "admin",
			Birthday:  &types.Date{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
		}

		actual := new(ID)
		paramName := "id"
		queryParams := url.Values{
			"id[firstName]": {"Alex"},
			"id[role]":      {"admin"},
			"foo":           {"bar"},
			"id[birthday]":  {"2020-01-01"},
		}

		err := BindQueryParameter("deepObject", true, false, paramName, queryParams, &actual)
		assert.NoError(t, err)
		assert.Equal(t, expectedDeepObject, actual)
	})

	t.Run("form", func(t *testing.T) {
		expected := &types.Date{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}
		birthday := &types.Date{}
		queryParams := url.Values{
			"birthday": {"2020-01-01"},
		}
		err := BindQueryParameter("form", true, false, "birthday", queryParams, &birthday)
		assert.NoError(t, err)
		assert.Equal(t, expected, birthday)
	})
}

func TestBindParameterViaAlias(t *testing.T) {
	// We don't need to check every parameter format type here, since the binding
	// code is identical irrespective of parameter type, buy we do want to test
	// a bunch of types.
	type AString string
	type Aint int
	type Afloat float64
	type Atime time.Time
	type Adate types.Date

	type AliasTortureTest struct {
		S  AString  `json:"s"`
		Sp *AString `json:"sp,omitempty"`
		I  Aint     `json:"i"`
		Ip *Aint    `json:"ip,omitempty"`
		F  Afloat   `json:"f"`
		Fp *Afloat  `json:"fp,omitempty"`
		T  Atime    `json:"t"`
		Tp *Atime   `json:"tp,omitempty"`
		D  Adate    `json:"d"`
		Dp *Adate   `json:"dp,omitempty"`
	}

	now := time.Now().UTC()
	later := now.Add(time.Hour)

	queryParams := url.Values{
		"alias[s]":  {"str"},
		"alias[sp]": {"strp"},
		"alias[i]":  {"1"},
		"alias[ip]": {"2"},
		"alias[f]":  {"3.5"},
		"alias[fp]": {"4.5"},
		"alias[t]":  {now.Format(time.RFC3339Nano)},
		"alias[tp]": {later.Format(time.RFC3339Nano)},
		"alias[d]":  {"2020-11-06"},
		"alias[dp]": {"2020-11-07"},
	}

	dst := new(AliasTortureTest)

	err := BindQueryParameter("deepObject", true, false, "alias", queryParams, &dst)
	require.NoError(t, err)

	var sp AString = "strp"
	var ip Aint = 2
	var fp Afloat = 4.5
	dp := Adate{Time: time.Date(2020, 11, 7, 0, 0, 0, 0, time.UTC)}

	expected := AliasTortureTest{
		S:  "str",
		Sp: &sp,
		I:  1,
		Ip: &ip,
		F:  3.5,
		Fp: &fp,
		T:  Atime(now),
		Tp: (*Atime)(&later),
		D:  Adate{Time: time.Date(2020, 11, 6, 0, 0, 0, 0, time.UTC)},
		Dp: &dp,
	}

	// Compare field by field, makes errors easier to track.
	assert.EqualValues(t, expected.S, dst.S)
	assert.EqualValues(t, expected.Sp, dst.Sp)
	assert.EqualValues(t, expected.I, dst.I)
	assert.EqualValues(t, expected.Ip, dst.Ip)
	assert.EqualValues(t, expected.F, dst.F)
	assert.EqualValues(t, expected.Fp, dst.Fp)
	assert.EqualValues(t, expected.T, dst.T)
	assert.EqualValues(t, expected.Tp, dst.Tp)
	assert.EqualValues(t, expected.D, dst.D)
	assert.EqualValues(t, expected.Dp, dst.Dp)
}

// bindParamsToExplodedObject has to special case some types. Make sure that
// these non-object types are handled correctly. The other parts of the functionality
// are tested via more generic code above.
func TestBindParamsToExplodedObject(t *testing.T) {
	now := time.Now().UTC()
	values := url.Values{
		"time": {now.Format(time.RFC3339Nano)},
		"date": {"2020-11-06"},
	}

	var dstTime time.Time
	err := bindParamsToExplodedObject("time", values, &dstTime)
	assert.NoError(t, err)
	assert.EqualValues(t, now, dstTime)

	type AliasedTime time.Time
	var aDstTime AliasedTime
	err = bindParamsToExplodedObject("time", values, &aDstTime)
	assert.NoError(t, err)
	assert.EqualValues(t, now, aDstTime)

	var dstDate types.Date
	expectedDate := types.Date{time.Date(2020, 11, 6, 0, 0, 0, 0, time.UTC)}
	err = bindParamsToExplodedObject("date", values, &dstDate)
	assert.EqualValues(t, expectedDate, dstDate)

	type AliasedDate types.Date
	var aDstDate AliasedDate
	err = bindParamsToExplodedObject("date", values, &aDstDate)
	assert.EqualValues(t, expectedDate, aDstDate)

}
