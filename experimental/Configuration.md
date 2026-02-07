# Configuration Reference

`oapi-codegen` is configured using a YAML file. All sections are optional — you only need to include what you want to change from the defaults.

Below is a fully annotated configuration file showing every option.

```yaml
# The Go package name for generated code.
# Can also be set with -package flag.
package: myapi

# Output file path.
# Can also be set with -output flag.
# Default: <spec-basename>.gen.go
output: types.gen.go

# Generation controls which parts of the code are generated.
generation:
  # Server framework to generate code for.
  # Supported: "std-http", "chi", "echo", "echo/v4", "gin", "gorilla", "fiber", "iris"
  # Default: "" (no server code generated)
  server: std-http

  # Generate an HTTP client that returns *http.Response.
  # Default: false
  client: true

  # Generate a SimpleClient wrapper with typed responses.
  # Requires client: true.
  # Default: false
  simple-client: true

  # Use model types from an external package instead of generating them locally.
  # When set, models are imported rather than generated.
  # Default: not set (models are generated locally)
  models-package:
    path: github.com/org/project/models
    alias: models  # optional, defaults to last segment of path

# Type mappings: OpenAPI type/format to Go type.
# User values are merged on top of defaults — you only need to specify overrides.
type-mapping:
  integer:
    default:
      type: int          # default
    formats:
      int32:
        type: int32      # default
      int64:
        type: int64      # default
  number:
    default:
      type: float32      # default
    formats:
      double:
        type: float64    # default
  boolean:
    default:
      type: bool         # default
  string:
    default:
      type: string       # default
    formats:
      byte:
        type: "[]byte"                         # default
      date:
        type: Date                             # default, custom template type
      date-time:
        type: time.Time                        # default
        import: time
      uuid:
        type: UUID                             # default, custom template type
      email:
        type: Email                            # default, custom template type
      binary:
        type: File                             # default, custom template type
      json:
        type: json.RawMessage                  # default
        import: encoding/json
      # Add your own format mappings:
      money:
        type: decimal.Decimal
        import: github.com/shopspring/decimal

# Name mangling: controls how OpenAPI names become Go identifiers.
# User values are merged on top of defaults.
name-mangling:
  # Prefix prepended when a name starts with a digit.
  # Default: "N" (e.g., "123foo" becomes "N123foo")
  numeric-prefix: "N"

  # Prefix prepended when a name conflicts with a Go keyword.
  # Default: "" (uses keyword-suffix instead)
  keyword-prefix: ""

  # Characters that mark word boundaries (next letter is capitalised).
  # Default includes most punctuation: - # @ ! $ & = . + : ; _ ~ space ( ) { } [ ] | < > ? / \
  word-separators: "-#@!$&=.+:;_~ (){}[]|<>?/\\"

  # Words that should remain all-uppercase.
  initialisms:
    - ACL
    - API
    - ASCII
    - CPU
    - CSS
    - DB
    - DNS
    - EOF
    - GUID
    - HTML
    - HTTP
    - HTTPS
    - ID
    - IP
    - JSON
    - QPS
    - RAM
    - RPC
    - SLA
    - SMTP
    - SQL
    - SSH
    - TCP
    - TLS
    - TTL
    - UDP
    - UI
    - UID
    - GID
    - URI
    - URL
    - UTF8
    - UUID
    - VM
    - XML
    - XMPP
    - XSRF
    - XSS
    - SIP
    - RTP
    - AMQP
    - TS

  # Characters that get replaced with words when they appear at the start of a name.
  character-substitutions:
    "$":  DollarSign
    "-":  Minus
    "+":  Plus
    "&":  And
    "|":  Or
    "~":  Tilde
    "=":  Equal
    ">":  GreaterThan
    "<":  LessThan
    "#":  Hash
    ".":  Dot
    "*":  Asterisk
    "^":  Caret
    "%":  Percent
    "_":  Underscore
    "@":  At
    "!":  Bang
    "?":  Question
    "/":  Slash
    "\\": Backslash
    ":":  Colon
    ";":  Semicolon
    "'":  Apos
    "\"": Quote
    "`":  Backtick
    "(":  LParen
    ")":  RParen
    "[":  LBracket
    "]":  RBracket
    "{":  LBrace
    "}":  RBrace

# Name substitutions: direct overrides for specific generated names.
name-substitutions:
  # Override type names during generation.
  type-names:
    foo: MyCustomFoo  # Schema "foo" generates type "MyCustomFoo" instead of "Foo"
  # Override property/field names during generation.
  property-names:
    bar: MyCustomBar  # Property "bar" generates field "MyCustomBar" instead of "Bar"

# Import mapping: resolve external $ref targets to Go packages.
# Required when your spec references schemas from other files.
import-mapping:
  ../common/api.yaml: github.com/org/project/common
  https://example.com/specs/shared.yaml: github.com/org/shared
  # Use "-" to indicate types should stay in the current package
  ./local-types.yaml: "-"

# Content types: regexp patterns controlling which media types generate models.
# Only request/response bodies with matching content types will have types generated.
# Default: JSON types only.
content-types:
  - "^application/json$"
  - "^application/.*\\+json$"
  # Add custom patterns as needed:
  # - "^application/xml$"
  # - "^text/plain$"

# Content type short names: maps content type patterns to short names
# used in generated type names (e.g., "FindPetsJSONResponse").
content-type-short-names:
  - pattern: "^application/json$"
    short-name: JSON
  - pattern: "^application/xml$"
    short-name: XML
  - pattern: "^text/plain$"
    short-name: Text
  # ... defaults cover JSON, XML, Text, HTML, Binary, Multipart, Form

# Struct tags: controls which struct tags are generated and their format.
# Uses Go text/template syntax. If specified, completely replaces the defaults.
# Default: json and form tags.
struct-tags:
  tags:
    - name: json
      template: '{{if .JSONIgnore}}-{{else}}{{ .FieldName }}{{if .OmitEmpty}},omitempty{{end}}{{if .OmitZero}},omitzero{{end}}{{end}}'
    - name: form
      template: '{{if .JSONIgnore}}-{{else}}{{ .FieldName }}{{if .OmitEmpty}},omitempty{{end}}{{end}}'
    # Add additional tags:
    - name: yaml
      template: '{{ .FieldName }}{{if .OmitEmpty}},omitempty{{end}}'
    - name: db
      template: '{{ .FieldName }}'
```

## Struct tag template variables

The struct tag templates have access to the following fields:

| Variable | Type | Description |
|----------|------|-------------|
| `.FieldName` | `string` | The original field name from the OpenAPI spec |
| `.GoFieldName` | `string` | The Go identifier name after name mangling |
| `.IsOptional` | `bool` | Whether the field is optional in the schema |
| `.IsNullable` | `bool` | Whether the field is nullable |
| `.IsPointer` | `bool` | Whether the field is rendered as a pointer type |
| `.OmitEmpty` | `bool` | Whether `omitempty` should be applied (from extensions or optionality) |
| `.OmitZero` | `bool` | Whether `omitzero` should be applied (from `x-oapi-codegen-omitzero` extension) |
| `.JSONIgnore` | `bool` | Whether the field should be ignored in JSON (from `x-go-json-ignore` extension) |
