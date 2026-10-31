// Package overlay tests applying OpenAPI Overlays. The rfc9535 and
// legacy_jsonpath packages are generated from spec.yaml with overlays that
// express the same actions, one with RFC 9535 targets and one with targets in
// the JSONPath dialect of oapi-codegen before v2.9.0, read through
// compatibility.overlays-use-yaml-jsonpath. See #2502.
package overlay

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	googleuuid "github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/overlay/legacy_jsonpath"
	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/overlay/rfc9535"
)

func TestBothDialectsProduceTheSameDocument(t *testing.T) {
	rfc, err := rfc9535.GetSpec()
	require.NoError(t, err)
	legacy, err := legacyjsonpath.GetSpec()
	require.NoError(t, err)

	rfcJSON, err := json.Marshal(rfc)
	require.NoError(t, err)
	legacyJSON, err := json.Marshal(legacy)
	require.NoError(t, err)
	assert.JSONEq(t, string(rfcJSON), string(legacyJSON))
}

func TestOverlayEffects(t *testing.T) {
	for name, getSpec := range map[string]func() (*openapi3.T, error){
		"rfc9535":         rfc9535.GetSpec,
		"legacy_jsonpath": legacyjsonpath.GetSpec,
	} {
		t.Run(name, func(t *testing.T) {
			spec, err := getSpec()
			require.NoError(t, err)

			// The structural update of $.
			assert.Equal(t, "structured-overlay", spec.Info.Extensions["x-overlay-applied"])
			ping := spec.Paths.Find("/ping")
			require.NotNil(t, ping)
			require.NotNil(t, ping.Get)
			assert.Equal(t, "Perform a ping request", *ping.Get.Responses.Status(200).Value.Description)

			// Internal paths and operations, removed by x-internal on the
			// path item, on the operation, by tag and by operationId.
			assert.Nil(t, spec.Paths.Find("/healthz"), "x-internal path")
			assert.Nil(t, ping.Delete, "x-internal operation")
			assert.Empty(t, spec.Paths.Find("/admin/autoscaling").Operations(), "operation tagged internal")
			assert.Empty(t, spec.Paths.Find("/debug/vars").Operations(), "operation matched by operationId")

			// Internal schemas and properties.
			schemas := spec.Components.Schemas
			assert.NotContains(t, schemas, "InternalAudit", "x-internal schema")
			assert.NotContains(t, schemas["Pong"].Value.Properties, "verbose", "x-internal schema property")
			assert.NotContains(t, schemas["User"].Value.Properties, "password_hash", "x-internal schema property")
			body := spec.Paths.Find("/users").Post.RequestBody.Value.Content.Get("application/json").Schema.Value
			assert.NotContains(t, body.Properties, "not_documented", "x-internal request body property")
			assert.Contains(t, body.Properties, "username")
			status := spec.Paths.Find("/status").Get.Responses.Status(200).Value.Content.Get("application/json").Schema.Value
			assert.NotContains(t, status.Properties, "build", "x-internal response property")
			assert.Contains(t, status.Properties, "healthy")

			// The renamed type and the removed property.
			assert.Equal(t, "OverriddenPong", schemas["Pong"].Value.Extensions["x-go-name"])
			assert.NotContains(t, schemas["Pong"].Value.Properties, "seed")
			assert.Equal(t, "PONG", schemas["Pong"].Value.Properties["ping"].Value.Example, "replaced scalar")
			// Numbers the overlay does not touch keep their values.
			ratio := schemas["Pong"].Value.Properties["ratio"].Value
			assert.Equal(t, 0.10, *ratio.Min)
			assert.Equal(t, 1.50, *ratio.Max)

			// Servers are added where an operation has none, and kept where
			// it has its own.
			require.Len(t, *ping.Get.Servers, 1)
			assert.Equal(t, "http://localhost:35123", (*ping.Get.Servers)[0].URL)
			statusServers := *spec.Paths.Find("/status").Get.Servers
			require.Len(t, statusServers, 1)
			assert.Equal(t, "https://status.example.com", statusServers[0].URL)

			// Merging into an operation and into a member whose name needs
			// brackets.
			assert.Equal(t, "Check that the API is running OK", ping.Get.Description)
			assert.Equal(t, "json", ping.Get.Responses.Status(200).Value.Content.Get("application/json").Extensions["x-media-type"])

			// Appending to sequences.
			assert.Equal(t, []string{"public", "ping"}, ping.Get.Tags)
			require.Len(t, spec.Tags, 2)
			assert.Equal(t, "ping", spec.Tags[1].Name)
			assert.Equal(t, "Pinging the API", spec.Tags[1].Description)
		})
	}
}

// TestGeneratedCode checks the overlay's effects on the generated code.
func TestGeneratedCode(t *testing.T) {
	for name, iface := range map[string]reflect.Type{
		"rfc9535":         reflect.TypeFor[rfc9535.ServerInterface](),
		"legacy_jsonpath": reflect.TypeFor[legacyjsonpath.ServerInterface](),
	} {
		t.Run(name, func(t *testing.T) {
			var methods []string
			for i := range iface.NumMethod() {
				methods = append(methods, iface.Method(i).Name)
			}
			assert.ElementsMatch(t, []string{"CreateUser", "ListWidgets", "Ping", "Status"}, methods, "only the operations the overlay keeps")
		})
	}

	// The renamed type, without the removed property or the internal one.
	for _, pong := range []reflect.Type{reflect.TypeFor[rfc9535.OverriddenPong](), reflect.TypeFor[legacyjsonpath.OverriddenPong]()} {
		var fields []string
		for i := range pong.NumField() {
			fields = append(fields, pong.Field(i).Name)
		}
		assert.Equal(t, []string{"Ping", "Ratio"}, fields)
	}
	_ = rfc9535.User{Username: "u"}
	_ = legacyjsonpath.User{Username: "u"}
}

// TestExtensionsAddedByTheOverlay checks that the extensions the overlays add,
// some through filtered targets, reach the generated code.
func TestExtensionsAddedByTheOverlay(t *testing.T) {
	// x-go-name on schemas, from a named target (Pong) and a filtered one
	// (Gizmo), and on properties, from a named target (label) and a filtered
	// one (serial). These only compile if the names were applied.
	rfc := rfc9535.Widget{
		Title:        "a widget",
		SerialNumber: &openapi_types.UUID{},
		RenamedGizmo: &rfc9535.RenamedGizmo{},
	}
	legacy := legacyjsonpath.Widget{
		Title:        "a widget",
		SerialNumber: &openapi_types.UUID{},
		RenamedGizmo: &legacyjsonpath.RenamedGizmo{},
	}
	_ = rfc9535.OverriddenPong{}
	_ = legacyjsonpath.OverriddenPong{}

	// x-go-type with x-go-type-import, from a filtered target: the schema is
	// the overriding type, imported under its alias.
	id := googleuuid.New()
	rfc.Id = id
	legacy.Id = id
	_, ok := any(rfc9535.WidgetIdentifier{}).(googleuuid.UUID)
	assert.True(t, ok, "WidgetIdentifier must be googleuuid.UUID")
	_, ok = any(legacyjsonpath.WidgetIdentifier{}).(googleuuid.UUID)
	assert.True(t, ok, "WidgetIdentifier must be googleuuid.UUID")

	// x-go-type-skip-optional-pointer, from a filtered target: the optional
	// date-time is a value.
	rfc.Created = time.Time{}
	legacy.Created = time.Time{}

	for name, widget := range map[string]reflect.Type{
		"rfc9535":         reflect.TypeOf(rfc),
		"legacy_jsonpath": reflect.TypeOf(legacy),
	} {
		t.Run(name, func(t *testing.T) {
			field := func(name string) reflect.StructField {
				t.Helper()
				f, ok := widget.FieldByName(name)
				require.True(t, ok, "field %s", name)
				return f
			}
			assert.Equal(t, reflect.TypeFor[time.Time](), field("Created").Type, "x-go-type-skip-optional-pointer")
			assert.Equal(t, reflect.TypeFor[googleuuid.UUID](), field("Id").Type, "x-go-type")
			assert.Equal(t, `json:"notes"`, string(field("Notes").Tag), "x-omitempty: false")
			assert.Equal(t, `json:"-"`, string(field("Secret").Tag), "x-go-json-ignore")
			assert.Equal(t, `json:"label"`, string(field("Title").Tag), "x-go-name keeps the JSON name")
			assert.Equal(t, `json:"serial,omitempty"`, string(field("SerialNumber").Tag), "x-go-name keeps the JSON name")
			// x-order puts zeta and alpha first, in that order.
			assert.Equal(t, "Zeta", widget.Field(0).Name)
			assert.Equal(t, "Alpha", widget.Field(1).Name)
		})
	}
}
