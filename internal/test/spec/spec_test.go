package spec

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestSwaggerUIHTML(t *testing.T) {
	swaggerUI, err := GetSwaggerUI("/swagger/spec.json")
	assert.Equal(t, err, nil)
	cupaloy.SnapshotT(t, swaggerUI)
}
