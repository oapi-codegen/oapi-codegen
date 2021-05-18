package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	InitVendorJSONRegex(DefaultVendorJSONRegex)
}

func TestIsJSON(t *testing.T) {
	assert.True(t, isJSON("application/json"))
	assert.False(t, isJSON("text/html"))
}

func TestIsVendorJSON(t *testing.T) {
	assert.True(t, isVendorJSON("application/vnd.company+json"))
	assert.False(t, isVendorJSON("application/json"))
}

func TestGetTagForVendorJSON(t *testing.T) {
	assert.Equal(t, "CompanyV1", getTagForVendorJSON("application/vnd.company.v1+json"))
	assert.Equal(t, "", getTagForVendorJSON("application/json"))
}

func TestGetVendorJSONTypeName(t *testing.T) {
	assert.Equal(t, "BodyCompanyV1", getVendorJSONTypeName("Body", "application/vnd.company.v1+json"))
}
