package issue1306_test

import (
	issue1306 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-1306"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestIssue1306(t *testing.T) {
	actualType := reflect.TypeOf(issue1306.SomeType{})
	expectedType := reflect.TypeOf(struct {
		Name    *string                    `json:"name,omitempty"`
		Objects *[]issue1306.SomeChildType `json:"objects,omitempty"`
	}{})

	assert.True(t, actualType.AssignableTo(expectedType))

	actualType = reflect.TypeOf(issue1306.SomeChildType{})
	expectedType = reflect.TypeOf(struct {
		Key   *string `json:"key,omitempty"`
		Value *string `json:"value,omitempty"`
	}{})

	assert.True(t, actualType.AssignableTo(expectedType))
}
