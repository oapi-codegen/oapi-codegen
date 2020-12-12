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

func TestBindStringToObject(t *testing.T) {
	var i int
	assert.NoError(t, BindStringToObject("5", &i))
	assert.Equal(t, 5, i)

	// Let's make sure we error out on things that can't be the correct
	// type. Since we're using reflect package setters, we'll have similar
	// unassignable type errors.
	assert.Error(t, BindStringToObject("5.7", &i))
	assert.Error(t, BindStringToObject("foo", &i))
	assert.Error(t, BindStringToObject("1,2,3", &i))

	var i8 int8
	assert.NoError(t, BindStringToObject("12", &i8))
	assert.Equal(t, int8(12), i8)

	assert.Error(t, BindStringToObject("5.7", &i8))
	assert.Error(t, BindStringToObject("foo", &i8))
	assert.Error(t, BindStringToObject("1,2,3", &i8))

	var i16 int16
	assert.NoError(t, BindStringToObject("12", &i16))
	assert.Equal(t, int16(12), i16)

	assert.Error(t, BindStringToObject("5.7", &i16))
	assert.Error(t, BindStringToObject("foo", &i16))
	assert.Error(t, BindStringToObject("1,2,3", &i16))

	var i32 int32
	assert.NoError(t, BindStringToObject("12", &i32))
	assert.Equal(t, int32(12), i32)

	assert.Error(t, BindStringToObject("5.7", &i32))
	assert.Error(t, BindStringToObject("foo", &i32))
	assert.Error(t, BindStringToObject("1,2,3", &i32))

	var i64 int64
	assert.NoError(t, BindStringToObject("124", &i64))
	assert.Equal(t, int64(124), i64)

	assert.Error(t, BindStringToObject("5.7", &i64))
	assert.Error(t, BindStringToObject("foo", &i64))
	assert.Error(t, BindStringToObject("1,2,3", &i64))

	var u uint
	assert.NoError(t, BindStringToObject("5", &u))
	assert.Equal(t, uint(5), u)

	assert.Error(t, BindStringToObject("5.7", &u))
	assert.Error(t, BindStringToObject("foo", &u))
	assert.Error(t, BindStringToObject("1,2,3", &u))

	var u8 uint8
	assert.NoError(t, BindStringToObject("12", &u8))
	assert.Equal(t, uint8(12), u8)

	assert.Error(t, BindStringToObject("5.7", &u8))
	assert.Error(t, BindStringToObject("foo", &u8))
	assert.Error(t, BindStringToObject("1,2,3", &u8))

	var u16 uint16
	assert.NoError(t, BindStringToObject("12", &u16))
	assert.Equal(t, uint16(12), u16)

	assert.Error(t, BindStringToObject("5.7", &u16))
	assert.Error(t, BindStringToObject("foo", &u16))
	assert.Error(t, BindStringToObject("1,2,3", &u16))

	var u32 uint32
	assert.NoError(t, BindStringToObject("12", &u32))
	assert.Equal(t, uint32(12), u32)

	assert.Error(t, BindStringToObject("5.7", &u32))
	assert.Error(t, BindStringToObject("foo", &u32))
	assert.Error(t, BindStringToObject("1,2,3", &u32))

	var u64 uint64
	assert.NoError(t, BindStringToObject("124", &u64))
	assert.Equal(t, uint64(124), u64)

	assert.Error(t, BindStringToObject("5.7", &u64))
	assert.Error(t, BindStringToObject("foo", &u64))
	assert.Error(t, BindStringToObject("1,2,3", &u64))

	var b bool
	assert.NoError(t, BindStringToObject("True", &b))
	assert.Equal(t, true, b)
	assert.NoError(t, BindStringToObject("true", &b))
	assert.Equal(t, true, b)
	assert.NoError(t, BindStringToObject("1", &b))
	assert.Equal(t, true, b)

	var f64 float64
	assert.NoError(t, BindStringToObject("1.25", &f64))
	assert.Equal(t, float64(1.25), f64)

	assert.Error(t, BindStringToObject("foo", &f64))
	assert.Error(t, BindStringToObject("1,2,3", &f64))

	var f32 float32
	assert.NoError(t, BindStringToObject("3.125", &f32))
	assert.Equal(t, float32(3.125), f32)

	assert.Error(t, BindStringToObject("foo", &f32))
	assert.Error(t, BindStringToObject("1,2,3", &f32))

	// This checks whether binding works through a type alias.
	type SomeType int
	var st SomeType
	assert.NoError(t, BindStringToObject("5", &st))
	assert.Equal(t, SomeType(5), st)

	// Check time binding
	now := time.Now().UTC()
	strTime := now.Format(time.RFC3339Nano)
	var parsedTime time.Time
	assert.NoError(t, BindStringToObject(strTime, &parsedTime))
	parsedTime = parsedTime.UTC()
	assert.EqualValues(t, now, parsedTime)

	now = now.Truncate(time.Second)
	strTime = now.Format(time.RFC3339)
	assert.NoError(t, BindStringToObject(strTime, &parsedTime))
	parsedTime = parsedTime.UTC()
	assert.EqualValues(t, now, parsedTime)

	// Checks whether time binding works through a type alias.
	type AliasedTime time.Time
	var aliasedTime AliasedTime
	assert.NoError(t, BindStringToObject(strTime, &aliasedTime))
	assert.EqualValues(t, now, aliasedTime)

	// Checks whether date binding works directly and through an alias.
	dateString := "2020-11-05"
	var dstDate types.Date
	assert.NoError(t, BindStringToObject(dateString, &dstDate))
	type AliasedDate types.Date
	var dstAliasedDate AliasedDate
	assert.NoError(t, BindStringToObject(dateString, &dstAliasedDate))

	// Checks whether a mock binder works and embedded types
	var mockBinder MockBinder
	assert.NoError(t, BindStringToObject(dateString, &mockBinder))
	assert.EqualValues(t, dateString, mockBinder.Time.Format("2006-01-02"))
	var dstEmbeddedMockBinder EmbeddedMockBinder
	assert.NoError(t, BindStringToObject(dateString, &dstEmbeddedMockBinder))
	assert.EqualValues(t, dateString, dstEmbeddedMockBinder.Time.Format("2006-01-02"))
}
