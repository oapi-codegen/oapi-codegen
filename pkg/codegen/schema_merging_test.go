package codegen

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaMergingBehaviorValidate(t *testing.T) {
	for _, tc := range []struct {
		name    string
		compat  CompatibilityOptions
		invalid bool
	}{
		{name: "unset", compat: CompatibilityOptions{}},
		{name: "v1", compat: CompatibilityOptions{SchemaMergingBehavior: "v1"}},
		{name: "v2", compat: CompatibilityOptions{SchemaMergingBehavior: "v2"}},
		{name: "unknown", compat: CompatibilityOptions{SchemaMergingBehavior: "v2.9"}, invalid: true},
		{name: "v3", compat: CompatibilityOptions{SchemaMergingBehavior: "v3"}},
		{name: "not a version yet", compat: CompatibilityOptions{SchemaMergingBehavior: "v4"}, invalid: true},
		{name: "old-merge-schemas alone", compat: CompatibilityOptions{OldMergeSchemas: true}},
		{name: "old-merge-schemas with v1", compat: CompatibilityOptions{OldMergeSchemas: true, SchemaMergingBehavior: "v1"}},
		{name: "old-merge-schemas with v2", compat: CompatibilityOptions{OldMergeSchemas: true, SchemaMergingBehavior: "v2"}, invalid: true},
		{name: "old-merge-schemas with v3", compat: CompatibilityOptions{OldMergeSchemas: true, SchemaMergingBehavior: "v3"}, invalid: true},
		{name: "old-allof-sibling-merging with v1", compat: CompatibilityOptions{OldAllOfSiblingMerging: true, SchemaMergingBehavior: "v1"}},
		{name: "old-allof-sibling-merging with v2", compat: CompatibilityOptions{OldAllOfSiblingMerging: true, SchemaMergingBehavior: "v2"}},
		{name: "old-allof-sibling-merging with v3", compat: CompatibilityOptions{OldAllOfSiblingMerging: true, SchemaMergingBehavior: "v3"}, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			problems := tc.compat.Validate()
			if !tc.invalid {
				assert.Empty(t, problems)
				return
			}
			assert.Contains(t, problems, "schema-merging-behavior")
		})
	}
}

// TestConfigurationValidateReportsSchemaMerging: the problem reaches
// Configuration.Validate, which is what the CLI calls.
func TestConfigurationValidateReportsSchemaMerging(t *testing.T) {
	cfg := Configuration{PackageName: "api", Compatibility: CompatibilityOptions{SchemaMergingBehavior: "v4"}}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema-merging-behavior")
}

// TestSchemaMergerFor: every version maps to its own allOf merge.
func TestSchemaMergerFor(t *testing.T) {
	pointer := func(f any) uintptr { return reflect.ValueOf(f).Pointer() }
	assert.Equal(t, pointer(mergeSchemasV2), pointer(schemaMergerFor(SchemaMergingV2)))
	assert.Equal(t, pointer(mergeSchemasV3), pointer(schemaMergerFor(SchemaMergingV3)))
	assert.NotEqual(t, pointer(mergeSchemasV2), pointer(schemaMergerFor(SchemaMergingV1)), "v1 has a merge of its own")
	assert.Panics(t, func() { schemaMergerFor("v4") })
}

func TestSchemaMergingVersion(t *testing.T) {
	for _, tc := range []struct {
		compat CompatibilityOptions
		want   string
	}{
		{CompatibilityOptions{}, SchemaMergingV2},
		{CompatibilityOptions{OldMergeSchemas: true}, SchemaMergingV1},
		{CompatibilityOptions{SchemaMergingBehavior: "v1"}, SchemaMergingV1},
		{CompatibilityOptions{SchemaMergingBehavior: "v2"}, SchemaMergingV2},
		{CompatibilityOptions{SchemaMergingBehavior: "v3"}, SchemaMergingV3},
	} {
		got, err := tc.compat.schemaMergingVersion()
		require.NoError(t, err)
		assert.Equal(t, tc.want, got, "%+v", tc.compat)
	}
}

// TestGenerateRejectsInvalidSchemaMerging: Generate refuses the settings
// Validate refuses, for callers that don't validate first.
func TestGenerateRejectsInvalidSchemaMerging(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
`
	for _, compat := range []CompatibilityOptions{
		{SchemaMergingBehavior: "v4"},
		{OldMergeSchemas: true, SchemaMergingBehavior: "v2"},
		{OldAllOfSiblingMerging: true, SchemaMergingBehavior: "v3"},
	} {
		_, err := generateSpecErr(spec, func(c *Configuration) { c.Compatibility = compat })
		require.Error(t, err, "%+v", compat)
		assert.Contains(t, err.Error(), "schema-merging-behavior")
	}
}

func TestOldMergeSchemasIsDeprecated(t *testing.T) {
	cfg := Configuration{PackageName: "api", Compatibility: CompatibilityOptions{OldMergeSchemas: true}}
	assert.Contains(t, cfg.Warnings()["old-merge-schemas"], "schema-merging-behavior: v1")
	cfg.Compatibility = CompatibilityOptions{SchemaMergingBehavior: "v1"}
	assert.NotContains(t, cfg.Warnings(), "old-merge-schemas")
}

// TestSchemaMergingV1IsOldMergeSchemas: v1 is what old-merge-schemas has
// always generated, and differs from v2 for an allOf.
//
// This test records how schema-merging-behavior v1 behaves. A bug fix may
// change it, but it must not be changed to accept a regression, and a commit
// that changes it must say why.
func TestSchemaMergingV1IsOldMergeSchemas(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Named:
      type: object
      properties:
        name: {type: string}
    Pet:
      allOf:
        - $ref: '#/components/schemas/Named'
        - type: object
          properties:
            kind: {type: string}
`
	alias := generateSpec(t, spec, func(c *Configuration) { c.Compatibility.OldMergeSchemas = true })
	v1 := generateSpec(t, spec, func(c *Configuration) { c.Compatibility.SchemaMergingBehavior = SchemaMergingV1 })
	v2 := generateSpec(t, spec, func(c *Configuration) { c.Compatibility.SchemaMergingBehavior = SchemaMergingV2 })
	assert.Equal(t, alias, v1)
	assert.Contains(t, v1, "Embedded struct due to allOf(#/components/schemas/Named)")
	assert.NotEqual(t, v1, v2)
	assert.Equal(t, generateSpec(t, spec), v2, "v2 is the default")
}
