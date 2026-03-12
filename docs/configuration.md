# Configuration Reference

`oapi-codegen` is configured using a YAML configuration file. This document describes every available setting.

You can also use [the JSON Schema](../configuration-schema.json) for IDE autocompletion:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: api
# ...
```

# Example configuration with all settings

Every setting is shown below with its default value. Click any setting name to jump to its documentation.

<pre>
# Required: Go package name
<a href="#package">package</a>: api

# What to generate (only one server type at a time)
<a href="#generate-options">generate</a>:
  <a href="#server-types">chi-server</a>: false
  <a href="#server-types">echo-server</a>: false
  <a href="#server-types">fiber-server</a>: false
  <a href="#server-types">gin-server</a>: false
  <a href="#server-types">gorilla-server</a>: false
  <a href="#server-types">iris-server</a>: false
  <a href="#server-types">std-http-server</a>: false
  <a href="#other-generate-flags">strict-server</a>: false
  <a href="#other-generate-flags">client</a>: false
  <a href="#other-generate-flags">models</a>: false
  <a href="#other-generate-flags">embedded-spec</a>: false
  <a href="#other-generate-flags">server-urls</a>: false

# Backward compatibility settings
<a href="#compatibility-options">compatibility</a>:
  <a href="#old-merge-schemas">old-merge-schemas</a>: false
  <a href="#old-enum-conflicts">old-enum-conflicts</a>: false
  <a href="#old-aliasing">old-aliasing</a>: false
  <a href="#disable-flatten-additional-properties">disable-flatten-additional-properties</a>: false
  <a href="#disable-required-readonly-as-pointer">disable-required-readonly-as-pointer</a>: false
  <a href="#always-prefix-enum-values">always-prefix-enum-values</a>: false
  <a href="#apply-chi-middleware-first-to-last">apply-chi-middleware-first-to-last</a>: false
  <a href="#apply-gorilla-middleware-first-to-last">apply-gorilla-middleware-first-to-last</a>: false
  <a href="#circular-reference-limit">circular-reference-limit</a>: 0
  <a href="#allow-unexported-struct-field-names">allow-unexported-struct-field-names</a>: false
  <a href="#preserve-original-operation-id-casing-in-embedded-spec">preserve-original-operation-id-casing-in-embedded-spec</a>: false

# Output modification options
<a href="#output-options">output-options</a>:
  <a href="#filtering">include-tags</a>: []
  <a href="#filtering">exclude-tags</a>: []
  <a href="#filtering">include-operation-ids</a>: []
  <a href="#filtering">exclude-operation-ids</a>: []
  <a href="#filtering">exclude-schemas</a>: []
  <a href="#formatting-and-pruning">skip-fmt</a>: false
  <a href="#formatting-and-pruning">skip-prune</a>: false
  <a href="#naming">name-normalizer</a>: ""
  <a href="#naming">initialism-overrides</a>: false
  <a href="#naming">additional-initialisms</a>: []
  <a href="#naming">response-type-suffix</a>: ""
  <a href="#naming">client-type-name</a>: ""
  <a href="#type-generation">nullable-type</a>: false
  <a href="#type-generation">disable-type-aliases-for-type</a>: []
  <a href="#type-generation">resolve-type-name-collisions</a>: false
  <a href="#type-generation">prefer-skip-optional-pointer</a>: false
  <a href="#type-generation">prefer-skip-optional-pointer-with-omitzero</a>: false
  <a href="#type-generation">prefer-skip-optional-pointer-on-container-types</a>: false
  <a href="#type-mapping">type-mapping</a>:
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
  <a href="#tags">yaml-tags</a>: false
  <a href="#client-options">client-response-bytes-function</a>: false
  <a href="#templates">user-templates</a>: {}
  <a href="#overlay">overlay</a>:
    <a href="#overlay">path</a>: ""
    <a href="#overlay">strict</a>: true

# External reference to Go package mapping
<a href="#import-mapping">import-mapping</a>: {}

# Additional Go imports
<a href="#additional-imports">additional-imports</a>: []
</pre>

# `package`

The Go package name for the generated code. This is the only required setting.

```yaml
package: api
```

# `import-mapping`

Maps external OpenAPI `$ref` paths to Go package import paths. Used when splitting specs across multiple packages.

```yaml
import-mapping:
  common.yaml: github.com/myorg/myproject/api/common
  auth.yaml: github.com/myorg/myproject/api/auth
```

# `additional-imports`

Defines additional Go imports to add to the generated code.

```yaml
additional-imports:
  - alias: mylib
    package: github.com/myorg/mylib
```

Each entry has:
- `package` (required): the Go import path
- `alias` (optional): an import alias

# Generate options

These are specified under the `generate` key. Only one server type may be enabled at a time.

If the `generate` block is omitted entirely, it defaults to generating an Echo server with models and an embedded spec.

## Server types

| Setting | Description |
| --- | --- |
| `chi-server` | Generate [Chi](https://github.com/go-chi/chi) server boilerplate |
| `echo-server` | Generate [Echo](https://github.com/labstack/echo) server boilerplate |
| `fiber-server` | Generate [Fiber](https://github.com/gofiber/fiber) server boilerplate |
| `gin-server` | Generate [Gin](https://github.com/gin-gonic/gin) server boilerplate |
| `gorilla-server` | Generate [gorilla/mux](https://github.com/gorilla/mux) server boilerplate |
| `iris-server` | Generate [Iris](https://github.com/kataras/iris) server boilerplate |
| `std-http-server` | Generate Go 1.22+ `net/http` server boilerplate |

## Other generate flags

| Setting | Description |
| --- | --- |
| `strict-server` | Generate a strict server wrapper. Must be used alongside one of the server types above. |
| `client` | Generate API client boilerplate |
| `models` | Generate Go type definitions from OpenAPI schemas |
| `embedded-spec` | Embed the OpenAPI spec in the generated code |
| `server-urls` | Generate types for the `servers` definitions' URLs |

# Compatibility options

These are specified under the `compatibility` key. They exist to preserve backward-compatible behavior when a bug fix or improvement changes generated output.

## `old-merge-schemas`

Reverts to the old behavior for `allOf` schema merging, which inlined each schema within the schema list rather than merging at the schema definition level. See [#531](https://github.com/oapi-codegen/oapi-codegen/issues/531).

## `old-enum-conflicts`

Reverts to old behavior for enum type naming that could produce conflicting typenames. See [#549](https://github.com/oapi-codegen/oapi-codegen/issues/549).

## `old-aliasing`

Reverts to generating a full Go type definition for every `$ref` instead of using type aliases where possible. See [#549](https://github.com/oapi-codegen/oapi-codegen/issues/549).

## `disable-flatten-additional-properties`

When an object contains no members and only an `additionalProperties` specification, it is normally flattened to a map. Set this to `true` to disable that behavior.

## `disable-required-readonly-as-pointer`

When an object property is both `required` and `readOnly`, the generated Go model uses a pointer by default. Set this to `true` to generate non-pointer types instead. See [#604](https://github.com/oapi-codegen/oapi-codegen/issues/604).

## `always-prefix-enum-values`

When set to `true`, always prefix enum constant variable names with their type name, even when there are no naming conflicts.

## `apply-chi-middleware-first-to-last`

Fixes the ordering of Chi middleware so they are chained in the order they are invoked, rather than the historically inverted order. See [#786](https://github.com/oapi-codegen/oapi-codegen/issues/786).

## `apply-gorilla-middleware-first-to-last`

Fixes the ordering of gorilla/mux middleware so they are chained in the order they are invoked, rather than the historically inverted order. See [#841](https://github.com/oapi-codegen/oapi-codegen/issues/841).

## `circular-reference-limit`

> **Deprecated**: In kin-openapi v0.126.0, the Circular Reference Counter functionality was removed, resolving all references with backtracking instead.

Allows controlling the limit for circular reference checking in older versions.

## `allow-unexported-struct-field-names`

Makes it possible to output structs that have unexported (lowercase) fields. Expected to be used in conjunction with `x-go-name`, `x-oapi-codegen-only-honour-go-name`, and `x-oapi-codegen-extra-tags`.

Example use case:

```yaml
# In your OpenAPI spec:
id:
  type: string
  x-go-name: accountIdentifier
  x-oapi-codegen-extra-tags:
    json: "-"
  x-oapi-codegen-only-honour-go-name: true
```

> [!NOTE]
> This can be confusing to users of your OpenAPI specification, who may see a field present and expect to use it in the request/response.

## `preserve-original-operation-id-casing-in-embedded-spec`

Ensures that the `operationId` from the source spec is kept intact in the embedded spec output, rather than having the `name-normalizer` applied to it. This keeps the embedded OpenAPI specification in sync with the input specification.

> [!NOTE]
> This does not impact generated code. If you're using `include-operation-ids` or `exclude-operation-ids`, ensure the `operationId`s used are correct.

# Output options

These are specified under the `output-options` key.

## Filtering

| Setting | Type | Description |
| --- | --- | --- |
| `include-tags` | `[]string` | Only include operations that have one of these tags. Ignored when empty. |
| `exclude-tags` | `[]string` | Exclude operations that have one of these tags. Ignored when empty. |
| `include-operation-ids` | `[]string` | Only include operations that have one of these operation IDs. Ignored when empty. |
| `exclude-operation-ids` | `[]string` | Exclude operations that have one of these operation IDs. Ignored when empty. |
| `exclude-schemas` | `[]string` | Exclude schemas with the given names from generation. Ignored when empty. |

## Formatting and pruning

| Setting | Type | Description |
| --- | --- | --- |
| `skip-fmt` | `bool` | Skip running `goimports` on the generated code |
| `skip-prune` | `bool` | Skip pruning unused components from the generated code |

## Naming

| Setting | Type | Description |
| --- | --- | --- |
| `name-normalizer` | `string` | Method used to normalize Go names and types. See [name normalizer values](#name-normalizer-values) below. |
| `initialism-overrides` | `bool` | Whether to use initialism overrides (e.g., `HTTP`, `API`) |
| `additional-initialisms` | `[]string` | Additional initialisms to recognize. Only has effect when `name-normalizer` is set to `ToCamelCaseWithInitialisms`. |
| `response-type-suffix` | `string` | Suffix used for response types |
| `client-type-name` | `string` | Override the default generated client type name |

### Name normalizer values

| Value | Example: `getHttpPet` | Example: `OneOf2things` |
| --- | --- | --- |
| (unset / `ToCamelCase`) | `GetHttpPet` | `OneOf2things` |
| `ToCamelCaseWithDigits` | `GetHttpPet` | `OneOf2Things` |
| `ToCamelCaseWithInitialisms` | `GetHTTPPet` | `OneOf2things` |

## Type generation

| Setting | Type | Description |
| --- | --- | --- |
| `nullable-type` | `bool` | Generate nullable types for nullable fields |
| `disable-type-aliases-for-type` | `[]string` | OpenAPI types that will not use type aliases. Currently supports: `"array"`. |
| `resolve-type-name-collisions` | `bool` | Automatically rename types that collide across different OpenAPI component sections by appending a suffix (e.g., `Parameter`, `Response`). Without this, codegen errors on duplicate type names. |
| `prefer-skip-optional-pointer` | `bool` | Globally omit the pointer for optional fields/types. Same as adding `x-go-type-skip-optional-pointer` to every field. |
| `prefer-skip-optional-pointer-with-omitzero` | `bool` | Generate the `omitzero` JSON tag for types that would have had an optional pointer. Must be used alongside `prefer-skip-optional-pointer`. A field can set `x-omitzero: false` to opt out. |
| `prefer-skip-optional-pointer-on-container-types` | `bool` | Disable the "optional pointer" for optional container types (slices, maps), avoiding unnecessary `!= nil` checks. |

## Type mapping

The `type-mapping` setting allows customizing how OpenAPI type/format combinations map to Go types. User-specified mappings are merged on top of the defaults.

```yaml
output-options:
  type-mapping:
    string:
      formats:
        date-time:
          type: mycustomtime.Time
          import: github.com/myorg/mycustomtime
    integer:
      default:
        type: int64
```

The structure is:

```yaml
type-mapping:
  <openapi-type>:       # integer, number, boolean, string
    default:
      type: <go-type>
      import: <go-import-path>   # optional
    formats:
      <format-name>:
        type: <go-type>
        import: <go-import-path> # optional
```

The default mappings are:

```yaml
type-mapping:
  integer:
    default:
      type: int
    formats:
      int:
        type: int
      int8:
        type: int8
      int16:
        type: int16
      int32:
        type: int32
      int64:
        type: int64
      uint:
        type: uint
      uint8:
        type: uint8
      uint16:
        type: uint16
      uint32:
        type: uint32
      uint64:
        type: uint64
  number:
    default:
      type: float32
    formats:
      float:
        type: float32
      double:
        type: float64
  boolean:
    default:
      type: bool
  string:
    default:
      type: string
    formats:
      byte:
        type: "[]byte"
      email:
        type: openapi_types.Email
      date:
        type: openapi_types.Date
      date-time:
        type: time.Time
        import: time
      json:
        type: json.RawMessage
        import: encoding/json
      uuid:
        type: openapi_types.UUID
      binary:
        type: openapi_types.File
```

## Tags

| Setting | Type | Description |
| --- | --- | --- |
| `yaml-tags` | `bool` | Add YAML struct tags to generated types, in addition to the default JSON tags |

## Client options

| Setting | Type | Description |
| --- | --- | --- |
| `client-response-bytes-function` | `bool` | Generate a `Bytes()` method on response objects for `ClientWithResponses` |

## Templates

| Setting | Type | Description |
| --- | --- | --- |
| `user-templates` | `map[string]string` | Override built-in templates with user-provided files. Keys are template names, values are file paths. |

## Overlay

The `overlay` setting configures an [OpenAPI Overlay](https://github.com/OAI/Overlay-Specification) to modify the spec before generation.

```yaml
output-options:
  overlay:
    path: overlay.yaml
    strict: true
```

| Setting | Type | Description |
| --- | --- | --- |
| `overlay.path` | `string` | Path to the overlay file |
| `overlay.strict` | `bool` | When `true` (the default), highlights any overlay actions that have no effect. Set to `false` to suppress these warnings. |
