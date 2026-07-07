# Configuration Reference

`oapi-codegen` is configured using a YAML configuration file.

This page is a structural overview. The authoritative, always-current reference for every setting is:

- the GoDoc for [`codegen.Configuration`](https://pkg.go.dev/github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen#Configuration), which documents each field alongside its YAML key
- [the JSON Schema](../configuration-schema.json), which gives your editor autocompletion, hover documentation and validation via the Language Server Protocol (LSP)

To get the IDE experience, add this line to the top of your configuration file:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: api
# ...
```

# Structure at a glance

Every setting is shown below with its default value. Each section links to the GoDoc describing its settings in detail.

<pre>
# Required: Go package name
package: api

# What to generate (only one server type at a time).
# If the `generate` block is omitted entirely, it defaults to generating
# an Echo server with models and an embedded spec.
# See <a href="https://pkg.go.dev/github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen#GenerateOptions">GenerateOptions</a>
generate:
  chi-server: false
  echo-server: false
  echo5-server: false      # requires Go 1.25+
  fiber-server: false
  gin-server: false
  gorilla-server: false
  iris-server: false
  std-http-server: false
  strict-server: false     # used alongside one of the server types above
  client: false
  models: false
  embedded-spec: false
  server-urls: false

# Backward compatibility settings. These preserve backward-compatible
# behavior when a bug fix or improvement changes generated output.
# See <a href="https://pkg.go.dev/github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen#CompatibilityOptions">CompatibilityOptions</a>
compatibility:
  old-merge-schemas: false
  old-allof-sibling-merging: false
  old-enum-conflicts: false
  old-aliasing: false
  disable-flatten-additional-properties: false
  disable-required-readonly-as-pointer: false
  always-prefix-enum-values: false
  apply-chi-middleware-first-to-last: false
  apply-gorilla-middleware-first-to-last: false
  circular-reference-limit: 0   # deprecated: no longer used
  allow-unexported-struct-field-names: false
  preserve-original-operation-id-casing-in-embedded-spec: false
  disable-enum-value-conflict-resolution: false
  headers-implicitly-required: false
  enable-auth-scopes-on-context: false

# Output modification options
# See <a href="https://pkg.go.dev/github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen#OutputOptions">OutputOptions</a>
output-options:
  include-tags: []
  exclude-tags: []
  include-operation-ids: []
  exclude-operation-ids: []
  exclude-schemas: []
  skip-fmt: false
  skip-prune: false
  name-normalizer: ""
  initialism-overrides: false
  additional-initialisms: []
  response-type-suffix: ""
  client-type-name: ""
  nullable-type: false
  disable-type-aliases-for-type: []
  resolve-type-name-collisions: false
  prefer-skip-optional-pointer: false
  prefer-skip-optional-pointer-with-omitzero: false
  prefer-skip-optional-pointer-on-container-types: false
  skip-enum-validate: false
  skip-enum-via-oneof: false
  generate-types-for-anonymous-schemas: false
  # How OpenAPI type/format combinations map to Go types; user-specified
  # mappings are merged on top of these defaults.
  type-mapping:
    integer:
      default:
        type: int
      formats:
        int:    { type: int }
        int8:   { type: int8 }
        int16:  { type: int16 }
        int32:  { type: int32 }
        int64:  { type: int64 }
        uint:   { type: uint }
        uint8:  { type: uint8 }
        uint16: { type: uint16 }
        uint32: { type: uint32 }
        uint64: { type: uint64 }
    number:
      default:
        type: float32
      formats:
        float:  { type: float32 }
        double: { type: float64 }
    boolean:
      default:
        type: bool
    string:
      default:
        type: string
      formats:
        byte:      { type: "[]byte" }
        email:     { type: openapi_types.Email }
        date:      { type: openapi_types.Date }
        date-time: { type: time.Time, import: time }
        json:      { type: json.RawMessage, import: encoding/json }
        uuid:      { type: openapi_types.UUID }
        binary:    { type: openapi_types.File }
  yaml-tags: false
  client-response-bytes-function: false
  skip-client-response-content-type: false
  skip-response-body-getters: false
  streaming-content-types: []
  user-templates: {}
  # OpenAPI Overlay applied to the spec before generation
  overlay:
    path: ""
    strict: true

# External reference to Go package mapping, used when
# <a href="../README.md#import-mapping">splitting specs across multiple packages</a>
import-mapping: {}

# Additional Go imports to add to the generated code
additional-imports: []
</pre>
