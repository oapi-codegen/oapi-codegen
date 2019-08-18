package spec

import (
	"fmt"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestSwaggerUIPage(t *testing.T) {
	swaggerUI, err := GetSwaggerUIPage("/swagger/spec.json", "/redirect.html")
	fmt.Println(err)
	assert.Equal(t, err, nil)
	cupaloy.SnapshotT(t, swaggerUI)
}

func TestSwaggerUIRedirect(t *testing.T) {
	redirectPage, err := GetSwaggerUIRedirect("/swagger/spec.json", "/redirect.html")
	fmt.Println(err)
	assert.Equal(t, err, nil)
	cupaloy.SnapshotT(t, redirectPage)
}
