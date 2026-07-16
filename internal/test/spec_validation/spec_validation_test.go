// Package spec_validation exercises codegen.ValidateSpec against specs that
// place a hostile value in a single sensitive field. Each testdata spec targets
// exactly one field, because ValidateSpec reports every problem it finds and we
// want each case to assert on a specific, friendly message. A separate spec
// proves that values which merely look suspicious but are legitimate (media-type
// parameters with quotes, multi-line descriptions) are accepted.
package spec_validation

import (
	"path/filepath"
	"testing"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSpecRejects(t *testing.T) {
	// file -> substring expected in the (friendly) error message.
	cases := map[string]string{
		"path_newline.yaml":           "path",
		"parameter_name_quote.yaml":   "parameter name",
		"property_name_backtick.yaml": "property name",
		"content_type_newline.yaml":   "content type",
		"enum_value_newline.yaml":     "enum value",
		"discriminator_quote.yaml":    "discriminator",
		"header_name_quote.yaml":      "header",
		"extra_tags_backtick.yaml":    "x-oapi-codegen-extra-tags",
		"x_go_name_invalid.yaml":      "x-go-name",
		"x_go_type_semicolon.yaml":    "x-go-type",
		"ref_sibling_x_go_type.yaml":  "x-go-type",
		"security_scope_quote.yaml":   "scope",

		"shared_path_param_type_collision.yaml": "path-level parameter",
	}

	for file, want := range cases {
		t.Run(file, func(t *testing.T) {
			spec, err := util.LoadSwagger(filepath.Join("testdata", file))
			require.NoError(t, err, "test spec should load")

			err = codegen.ValidateSpec(spec)
			require.Error(t, err, "ValidateSpec should reject %s", file)
			assert.Contains(t, err.Error(), want, "error message should point at the offending field")
		})
	}
}

func TestValidateSpecAcceptsLegitimateSpec(t *testing.T) {
	spec, err := util.LoadSwagger(filepath.Join("testdata", "valid.yaml"))
	require.NoError(t, err)
	require.NoError(t, codegen.ValidateSpec(spec), "a legitimate spec must not be rejected")
}
