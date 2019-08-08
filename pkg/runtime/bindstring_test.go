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
)

func unescapePassThrough(s string) (string, error) {
	return s, nil
}

func TestBindStringToObject(t *testing.T) {
	var s string
	assert.NoError(t, BindStringToObject(url.PathUnescape, "foo", &s))
	assert.Equal(t, "foo", s)
	assert.NoError(t, BindStringToObject(url.QueryUnescape, "foo", &s))
	assert.Equal(t, "foo", s)
	assert.NoError(t, BindStringToObject(unescapePassThrough, "foo", &s))
	assert.Equal(t, "foo", s)
	assert.NoError(t, BindStringToObject(url.PathUnescape, "foo%3Fbar", &s))
	assert.Equal(t, "foo?bar", s)
	assert.NoError(t, BindStringToObject(url.QueryUnescape, "foo%3Fbar", &s))
	assert.Equal(t, "foo?bar", s)
	assert.NoError(t, BindStringToObject(unescapePassThrough, "foo%3Fbar", &s))
	assert.Equal(t, "foo%3Fbar", s)
	assert.NoError(t, BindStringToObject(url.PathUnescape, "foo+bar", &s))
	assert.Equal(t, "foo+bar", s)
	assert.NoError(t, BindStringToObject(url.QueryUnescape, "foo+bar", &s))
	assert.Equal(t, "foo bar", s)
	assert.NoError(t, BindStringToObject(unescapePassThrough, "foo+bar", &s))
	assert.Equal(t, "foo+bar", s)

	var i int
	assert.NoError(t, BindStringToObject(url.PathUnescape, "5", &i))
	assert.Equal(t, 5, i)

	// Let's make sure we error out on things that can't be the correct
	// type. Since we're using reflect package setters, we'll have similar
	// unassignable type errors.
	assert.Error(t, BindStringToObject(url.PathUnescape, "5.7", &i))
	assert.Error(t, BindStringToObject(url.PathUnescape, "foo", &i))
	assert.Error(t, BindStringToObject(url.PathUnescape, "1,2,3", &i))

	var i64 int64
	assert.NoError(t, BindStringToObject(url.PathUnescape, "124", &i64))
	assert.Equal(t, int64(124), i64)

	assert.Error(t, BindStringToObject(url.PathUnescape, "5.7", &i64))
	assert.Error(t, BindStringToObject(url.PathUnescape, "foo", &i64))
	assert.Error(t, BindStringToObject(url.PathUnescape, "1,2,3", &i64))

	var i32 int32
	assert.NoError(t, BindStringToObject(url.PathUnescape, "12", &i32))
	assert.Equal(t, int32(12), i32)

	assert.Error(t, BindStringToObject(url.PathUnescape, "5.7", &i32))
	assert.Error(t, BindStringToObject(url.PathUnescape, "foo", &i32))
	assert.Error(t, BindStringToObject(url.PathUnescape, "1,2,3", &i32))

	var b bool
	assert.NoError(t, BindStringToObject(url.PathUnescape, "True", &b))
	assert.Equal(t, true, b)
	assert.NoError(t, BindStringToObject(url.PathUnescape, "true", &b))
	assert.Equal(t, true, b)
	assert.NoError(t, BindStringToObject(url.PathUnescape, "1", &b))
	assert.Equal(t, true, b)

	var f64 float64
	assert.NoError(t, BindStringToObject(url.PathUnescape, "1.25", &f64))
	assert.Equal(t, float64(1.25), f64)

	assert.Error(t, BindStringToObject(url.PathUnescape, "foo", &f64))
	assert.Error(t, BindStringToObject(url.PathUnescape, "1,2,3", &f64))

	var f32 float32
	assert.NoError(t, BindStringToObject(url.PathUnescape, "3.125", &f32))
	assert.Equal(t, float32(3.125), f32)

	assert.Error(t, BindStringToObject(url.PathUnescape, "foo", &f32))
	assert.Error(t, BindStringToObject(url.PathUnescape, "1,2,3", &f32))

	// This checks whether binding works through a type alias.
	type SomeType int
	var st SomeType
	assert.NoError(t, BindStringToObject(url.PathUnescape, "5", &st))
	assert.Equal(t, SomeType(5), st)

	// Check time binding
	now := time.Now().UTC()
	strTime := now.Format(time.RFC3339Nano)
	var parsedTime time.Time
	assert.NoError(t, BindStringToObject(url.PathUnescape, strTime, &parsedTime))
	parsedTime = parsedTime.UTC()
	assert.EqualValues(t, now, parsedTime)

	now = now.Truncate(time.Second)
	strTime = now.Format(time.RFC3339)
	assert.NoError(t, BindStringToObject(url.PathUnescape, strTime, &parsedTime))
	parsedTime = parsedTime.UTC()
	assert.EqualValues(t, now, parsedTime)
}
