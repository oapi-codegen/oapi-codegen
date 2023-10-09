package issue1306_test

import (
	issue1306 "github.com/deepmap/oapi-codegen/internal/test/issues/issue-1306"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestIssue1306(t *testing.T) {
	actualType := reflect.TypeOf(issue1306.SomeObject{})
	expectedType := reflect.TypeOf(struct {
		Name *string `json:"name,omitempty"`
	}{})

	assert.True(t, actualType.AssignableTo(expectedType))
}
