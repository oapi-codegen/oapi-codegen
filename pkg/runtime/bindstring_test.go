package runtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	var i64 int64
	assert.NoError(t, BindStringToObject("124", &i64))
	assert.Equal(t, int64(124), i64)

	assert.Error(t, BindStringToObject("5.7", &i64))
	assert.Error(t, BindStringToObject("foo", &i64))
	assert.Error(t, BindStringToObject("1,2,3", &i64))

	var i32 int32
	assert.NoError(t, BindStringToObject("12", &i32))
	assert.Equal(t, int32(12), i32)

	assert.Error(t, BindStringToObject("5.7", &i32))
	assert.Error(t, BindStringToObject("foo", &i32))
	assert.Error(t, BindStringToObject("1,2,3", &i32))

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
}
