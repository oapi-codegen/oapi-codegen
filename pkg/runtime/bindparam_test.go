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
	"math/big"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deepmap/oapi-codegen/pkg/types"
)

// MockBinder is just an independent version of Binder that has the Bind implemented
type MockBinder struct {
	time.Time
}

func (d *MockBinder) Bind(src string) error {
	// Don't fail on empty string.
	if src == "" {
		return nil
	}
	parsedTime, err := time.Parse(types.DateFormat, src)
	if err != nil {
		return fmt.Errorf("error parsing '%s' as date: %s", src, err)
	}
	d.Time = parsedTime
	return nil
}

func (d MockBinder) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format(types.DateFormat))
}

func (d *MockBinder) UnmarshalJSON(data []byte) error {
	var dateStr string
	err := json.Unmarshal(data, &dateStr)
	if err != nil {
		return err
	}
	parsed, err := time.Parse(types.DateFormat, dateStr)
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}

// EmbeddedMockBinder has an embedded MockBinder and so keeps the Binder Method from MockBinder.
type EmbeddedMockBinder struct {
	MockBinder
}

// AnotherMockBinder is an entirely new type we have to create a bind method with it to implement Binder as well.
type AnotherMockBinder MockBinder

func (b *AnotherMockBinder) Bind(src string) error {
	// Don't fail on empty string.
	if src == "" {
		return nil
	}
	parsedTime, err := time.Parse(types.DateFormat, src)
	if err != nil {
		return fmt.Errorf("error parsing '%s' as date: %s", src, err)
	}
	b.Time = parsedTime
	return nil
}

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
			Married   *MockBinder `json:"married"`
		}

		expectedName := "Alex"
		expectedDeepObject := &ID{
			FirstName: &expectedName,
			Role:      "admin",
			Birthday:  &types.Date{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
			Married:   &MockBinder{time.Date(2020, 2, 2, 0, 0, 0, 0, time.UTC)},
		}

		actual := new(ID)
		paramName := "id"
		queryParams := url.Values{
			"id[firstName]": {"Alex"},
			"id[role]":      {"admin"},
			"foo":           {"bar"},
			"id[birthday]":  {"2020-01-01"},
			"id[married]":   {"2020-02-02"},
		}

		err := BindQueryParameter("deepObject", true, false, paramName, queryParams, &actual)
		assert.NoError(t, err)
		assert.Equal(t, expectedDeepObject, actual)
	})

	t.Run("form", func(t *testing.T) {
		expected := &MockBinder{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}
		birthday := &MockBinder{}
		queryParams := url.Values{
			"birthday": {"2020-01-01"},
		}
		err := BindQueryParameter("form", true, false, "birthday", queryParams, &birthday)
		assert.NoError(t, err)
		assert.Equal(t, expected, birthday)
	})

	t.Run("optional", func(t *testing.T) {
		queryParams := url.Values{
			"time":   {"2020-12-09T16:09:53+00:00"},
			"number": {"100"},
		}
		// An optional time will be a pointer to a time in a parameter object
		var optionalTime *time.Time
		err := BindQueryParameter("form", true, false, "notfound", queryParams, &optionalTime)
		require.NoError(t, err)
		assert.Nil(t, optionalTime)

		var optionalNumber *int
		err = BindQueryParameter("form", true, false, "notfound", queryParams, &optionalNumber)
		require.NoError(t, err)
		assert.Nil(t, optionalNumber)

		// If we require values, we require errors when they're not present.
		err = BindQueryParameter("form", true, true, "notfound", queryParams, &optionalTime)
		assert.Error(t, err)
		err = BindQueryParameter("form", true, true, "notfound", queryParams, &optionalNumber)
		assert.Error(t, err)

	})
}

func TestBindParameterViaAlias(t *testing.T) {
	// We don't need to check every parameter format type here, since the binding
	// code is identical irrespective of parameter type, buy we do want to test
	// a bunch of types.
	type AString string
	type Aint int
	type Afloat float64
	type Atime = time.Time
	type Adate = MockBinder

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
		T:  now,
		Tp: &later,
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
	fieldsPresent, err := bindParamsToExplodedObject("time", values, &dstTime)
	assert.NoError(t, err)
	assert.True(t, fieldsPresent)
	assert.EqualValues(t, now, dstTime)

	type AliasedTime time.Time
	var aDstTime AliasedTime
	fieldsPresent, err = bindParamsToExplodedObject("time", values, &aDstTime)
	assert.NoError(t, err)
	assert.True(t, fieldsPresent)
	assert.EqualValues(t, now, aDstTime)

	expectedDate := MockBinder{Time: time.Date(2020, 11, 6, 0, 0, 0, 0, time.UTC)}

	var dstDate MockBinder
	fieldsPresent, err = bindParamsToExplodedObject("date", values, &dstDate)
	assert.NoError(t, err)
	assert.True(t, fieldsPresent)
	assert.EqualValues(t, expectedDate, dstDate)

	var eDstDate EmbeddedMockBinder
	fieldsPresent, err = bindParamsToExplodedObject("date", values, &eDstDate)
	assert.NoError(t, err)
	assert.True(t, fieldsPresent)
	assert.EqualValues(t, expectedDate, dstDate)

	var nTDstDate AnotherMockBinder
	fieldsPresent, err = bindParamsToExplodedObject("date", values, &nTDstDate)
	assert.NoError(t, err)
	assert.True(t, fieldsPresent)
	assert.EqualValues(t, expectedDate, nTDstDate)

	type ObjectWithOptional struct {
		Time *time.Time `json:"time,omitempty"`
	}

	var optDstTime ObjectWithOptional
	fieldsPresent, err = bindParamsToExplodedObject("explodedObject", values, &optDstTime)
	assert.NoError(t, err)
	assert.True(t, fieldsPresent)
	assert.EqualValues(t, &now, optDstTime.Time)
}

func TestBindStyledParameterWithLocation(t *testing.T) {
	expectedBig := big.NewInt(12345678910)

	var dstBigNumber big.Int
	err := BindStyledParameterWithLocation("simple", false, "id", ParamLocationUndefined,
		"12345678910", &dstBigNumber)
	assert.NoError(t, err)
	assert.Equal(t, *expectedBig, dstBigNumber)
}
