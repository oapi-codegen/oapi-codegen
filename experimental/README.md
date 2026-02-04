# oapi-codegen experimental

Ground-up rewrite of oapi-codegen with:

- **OpenAPI 3.1/3.2 support** via [libopenapi](https://github.com/pb33f/libopenapi) (replacing kin-openapi)
- **Clean separation**: model generation as core, routers as separate layers
- **No runtime package** - generated code is self-contained
- **Proper anyOf/oneOf union types** with correct marshal/unmarshal semantics
- **allOf flattening** - properties merged into single struct, not embedding
- **Default values** - generated `ApplyDefaults()` methods

## Usage

```bash
go run ./cmd/oapi-codegen -package myapi -output types.gen.go openapi.yaml
```

Options:
- `-package <name>` - Go package name for generated code
- `-output <file>` - Output file path (default: `<spec-basename>.gen.go`)
- `-config <file>` - Path to configuration file

## Structure

```
experimental/
├── cmd/oapi-codegen/          # CLI tool
├── internal/codegen/
│   ├── codegen.go             # Main generation orchestration
│   ├── gather.go              # Schema discovery from OpenAPI doc
│   ├── schema.go              # SchemaDescriptor, SchemaPath types
│   ├── typegen.go             # Go type expression generation
│   ├── output.go              # Code construction utilities
│   ├── namemangling.go        # Go identifier conversion
│   ├── typemapping.go         # OpenAPI → Go type mappings
│   ├── templates/             # Custom type templates
│   │   └── types/             # Email, UUID, File types
│   └── test/                  # Test suites
│       ├── files/             # Test OpenAPI specs
│       ├── comprehensive/     # Full feature tests
│       ├── default_values/    # ApplyDefaults tests
│       └── nested_aggregate/  # Union nesting tests
```

## Features

### Type Generation

| OpenAPI | Go |
|---------|-----|
| `type: object` with properties | `struct` |
| `type: object` with only `additionalProperties` | `map[string]T` |
| `type: object` with both | `struct` with `AdditionalProperties` field |
| `enum` | `type T string` with `const` block |
| `allOf` | Flattened `struct` (properties merged) |
| `anyOf` | Union `struct` with pointer fields + marshal/unmarshal |
| `oneOf` | Union `struct` with exactly-one enforcement |
| `$ref` | Resolved to target type name |

### Pointer Semantics

- **Required + non-nullable** → value type
- **Optional or nullable** → pointer type
- **Slices and maps** → never pointers (nil is zero value)

### Default Values

All structs get an `ApplyDefaults()` method:

```go
type Config struct {
    Name    *string `json:"name,omitempty"`
    Timeout *int    `json:"timeout,omitempty"`
}

func (s *Config) ApplyDefaults() {
    if s.Name == nil {
        v := "default-name"
        s.Name = &v
    }
    if s.Timeout == nil {
        v := 30
        s.Timeout = &v
    }
}
```

Usage pattern:
```go
var cfg Config
json.Unmarshal(data, &cfg)
cfg.ApplyDefaults()
```

### Union Types (anyOf/oneOf)

```go
// Generated for oneOf with two object variants
type MyUnion struct {
    VariantA *TypeA
    VariantB *TypeB
}

// MarshalJSON enforces exactly one variant set
// UnmarshalJSON tries each variant, expects exactly one match
```

## Configuration

Create a YAML configuration file:

```yaml
# config.yaml
package: myapi
output: types.gen.go

# Type mappings: OpenAPI type/format → Go type
type-mapping:
  integer:
    default:
      type: int
    formats:
      int64:
        type: int64
  number:
    default:
      type: float64  # Override default from float32
  string:
    formats:
      date-time:
        type: time.Time
        import: time
      uuid:
        type: uuid.UUID
        import: github.com/google/uuid
      # Custom format mapping
      money:
        type: decimal.Decimal
        import: github.com/shopspring/decimal

# Name mangling: controls how OpenAPI names become Go identifiers
name-mangling:
  # Prefix for names starting with digits
  numeric-prefix: "N"  # default: "N"
  # Prefix for Go reserved keywords
  reserved-prefix: ""  # default: "" (uses suffix instead)
  # Suffix for Go reserved keywords
  reserved-suffix: "_"  # default: "_"
  # Known initialisms (keep uppercase)
  initialisms:
    - ID
    - HTTP
    - URL
    - API
    - JSON
    - XML
    - UUID
  # Character substitutions
  character-substitutions:
    "+": "Plus"
    "@": "At"

# Name substitutions: direct overrides for specific names
name-substitutions:
  # Override individual schema/property names during conversion
  type-names:
    foo: MyCustomFoo  # "foo" becomes "MyCustomFoo" instead of "Foo"
  property-names:
    bar: MyCustomBar  # "bar" field becomes "MyCustomBar"
```

Use with `-config`:
```bash
go run ./cmd/oapi-codegen -config config.yaml openapi.yaml
```

### Default Type Mappings

| OpenAPI Type | Format | Go Type |
|--------------|--------|---------|
| `integer` | (none) | `int` |
| `integer` | `int32` | `int32` |
| `integer` | `int64` | `int64` |
| `number` | (none) | `float32` |
| `number` | `double` | `float64` |
| `boolean` | (none) | `bool` |
| `string` | (none) | `string` |
| `string` | `byte` | `[]byte` |
| `string` | `date-time` | `time.Time` |
| `string` | `uuid` | `UUID` (custom type) |
| `string` | `email` | `Email` (custom type) |
| `string` | `binary` | `File` (custom type) |

### Name Override Limitations

> **Note:** The current `name-substitutions` system only overrides individual name *parts* during conversion, not full generated type names.
>
> For example, if you have a schema at `#/components/schemas/Cat`:
> - Setting `type-names: {Cat: Kitty}` will produce `KittySchemaComponent` (stable) and `Kitty` (short)
> - You cannot currently override the full stable name `CatSchemaComponent` to something completely different
>
> Full type name overrides (by schema path or generated name) are not yet implemented.

## Development

Run tests:
```bash
go test ./internal/codegen/...
```

Regenerate test outputs:
```bash
go run ./cmd/oapi-codegen -package output -output internal/codegen/test/comprehensive/output/comprehensive.gen.go internal/codegen/test/files/comprehensive.yaml
```

## Status

**Active development.** Model generation is working for most schema types.

Working:
- [x] Object schemas → structs
- [x] Enum schemas → const blocks
- [x] allOf → flattened structs
- [x] anyOf/oneOf → union types
- [x] additionalProperties
- [x] $ref resolution
- [x] Nested/inline schemas
- [x] Default values (`ApplyDefaults()`)
- [x] Custom format types (email, uuid, binary)

Not yet implemented:
- [ ] Request/response body generation
- [ ] Operation/handler generation
- [ ] Router integrations
- [ ] Validation
- [ ] Client generation

See [CONTEXT.md](CONTEXT.md) for detailed design decisions and architecture notes.
