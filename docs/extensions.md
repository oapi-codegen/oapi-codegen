# OpenAPI Extensions

As well as the core OpenAPI support, `oapi-codegen` also supports the following OpenAPI extensions, as denoted by the [OpenAPI Specification Extensions](https://spec.openapis.org/oas/v3.0.3#specification-extensions).

## `x-go-type` / `x-go-type-import`

Override the generated type definition (and optionally, add an import from another package).

Using the `x-go-type` (and optionally, `x-go-type-import` when you need to import another package) allows overriding the type that `oapi-codegen` determined the generated type should be.

We can see this at play with the following schemas:

```yaml
components:
  schemas:
    Client:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
    ClientWithExtension:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          # this is a bit of a contrived example, as you could instead use
          # `format: uuid` but it explains how you'd do this when there may be
          # a clash, for instance if you already had a `uuid` package that was
          # being imported, or ...
          x-go-type: googleuuid.UUID
          x-go-type-import:
            path: github.com/google/uuid
            name: googleuuid
        id:
          type: number
          # ... this is also a bit of a contrived example, as you could use
          # `type: integer` but in the case that you know better than what
          # oapi-codegen is generating, like so:
          x-go-type: int64
```

From here, we now get two different models:

```go
// Client defines model for Client.
type Client struct {
	Id   *float32 `json:"id,omitempty"`
	Name string   `json:"name"`
}

// ClientWithExtension defines model for ClientWithExtension.
type ClientWithExtension struct {
	Id   *int64          `json:"id,omitempty"`
	Name googleuuid.UUID `json:"name"`
}
```

You can see this in more detail in [the example code](../examples/extensions/xgotype/).

## `x-go-type-skip-optional-pointer`

Do not add a pointer type for optional fields in structs.

> [!TIP]
> If you prefer this behaviour, and prefer to not have to annotate your whole OpenAPI spec for this behaviour, you can use `output-options.prefer-skip-optional-pointer=true` to default this behaviour for all fields.
>
> It is then possible to override this on a per-type/per-field basis where necessary.

By default, `oapi-codegen` will generate a pointer for optional fields.

Using the `x-go-type-skip-optional-pointer` extension allows omitting that pointer.

We can see this at play with the following schemas:

```yaml
components:
  schemas:
    Client:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
    ClientWithExtension:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
          x-go-type-skip-optional-pointer: true
```

From here, we now get two different models:

```go
// Client defines model for Client.
type Client struct {
	Id   *float32 `json:"id,omitempty"`
	Name string   `json:"name"`
}

// ClientWithExtension defines model for ClientWithExtension.
type ClientWithExtension struct {
	Id   float32 `json:"id,omitempty"`
	Name string  `json:"name"`
}
```

You can see this in more detail in [the example code](../examples/extensions/xgotypeskipoptionalpointer/).

## `x-go-name`

Override the generated name of a field or a type.

By default, `oapi-codegen` will attempt to generate the name of fields and types in as best a way it can.

However, sometimes, the name doesn't quite fit what your codebase standards are, or the intent of the field, so you can override it with `x-go-name`.

We can see this at play with the following schemas:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-go-name
components:
  schemas:
    Client:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
    ClientWithExtension:
      type: object
      # can be used on a type
      x-go-name: ClientRenamedByExtension
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
          # or on a field
          x-go-name: AccountIdentifier
```

From here, we now get two different models:

```go
// Client defines model for Client.
type Client struct {
	Id   *float32 `json:"id,omitempty"`
	Name string   `json:"name"`
}

// ClientRenamedByExtension defines model for ClientWithExtension.
type ClientRenamedByExtension struct {
	AccountIdentifier *float32 `json:"id,omitempty"`
	Name              string   `json:"name"`
}
```

You can see this in more detail in [the example code](../examples/extensions/xgoname/).

## `x-go-type-name`

Override the generated name of a type.

> [!NOTE]
> Notice that this is subtly different to the `x-go-name`, which also applies to _fields_ within `struct`s.

By default, `oapi-codegen` will attempt to generate the name of types in as best a way it can.

However, sometimes, the name doesn't quite fit what your codebase standards are, or the intent of the field, so you can override it with `x-go-name`.

We can see this at play with the following schemas:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-go-type-name
components:
  schemas:
    Client:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
    ClientWithExtension:
      type: object
      x-go-type-name: ClientRenamedByExtension
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
          # NOTE attempting a `x-go-type-name` here is a no-op, as we're not producing a _type_ only a _field_
          x-go-type-name: ThisWillNotBeUsed
```

From here, we now get two different models and a type alias:

```go
// Client defines model for Client.
type Client struct {
	Id   *float32 `json:"id,omitempty"`
	Name string   `json:"name"`
}

// ClientWithExtension defines model for ClientWithExtension.
type ClientWithExtension = ClientRenamedByExtension

// ClientRenamedByExtension defines model for .
type ClientRenamedByExtension struct {
	Id   *float32 `json:"id,omitempty"`
	Name string   `json:"name"`
}
```

You can see this in more detail in [the example code](../examples/extensions/xgotypename/).

## `x-omitempty`

Force the presence of the JSON tag `omitempty` on a field.

In a case that you may want to add the JSON struct tag `omitempty` to types that don't have one generated by default - for instance a required field - you can use the `x-omitempty` extension.

We can see this at play with the following schemas:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-omitempty
components:
  schemas:
    Client:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
    ClientWithExtension:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          # for some reason, you may want this behaviour, even though it's a required field
          x-omitempty: true
        id:
          type: number
```

From here, we now get two different models:

```go
// Client defines model for Client.
type Client struct {
	Id   *float32 `json:"id,omitempty"`
	Name string   `json:"name"`
}

// ClientWithExtension defines model for ClientWithExtension.
type ClientWithExtension struct {
	Id   *float32 `json:"id,omitempty"`
	Name string   `json:"name,omitempty"`
}
```

You can see this in more detail in [the example code](../examples/extensions/xomitempty/).

## `x-omitzero`

Force the presence of the JSON tag `omitzero` on a field.

> [!NOTE]
> `omitzero` was added in Go 1.24. If you're not using Go 1.24 in your project, this won't work.

In a case that you may want to add the JSON struct tag `omitzero` to types, you can use the `x-omitempty` extension.

We can see this at play with the following schemas:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-omitempty
components:
  schemas:
    Client:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
    ClientWithExtension:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
          x-omitzero: true
```

From here, we now get two different models:

```go
// Client defines model for Client.
type Client struct {
	Id   *float32 `json:"id,omitempty"`
	Name string   `json:"name"`
}

// ClientWithExtension defines model for ClientWithExtension.
type ClientWithExtension struct {
	Id   *float32 `json:"id,omitempty,omitzero"`
	Name string   `json:"name"`
}
```

You can see this in more detail in [the example code](../examples/extensions/xomitzero/).

## `x-go-json-ignore`

When (un)marshaling JSON, ignore field(s).

By default, `oapi-codegen` will generate `json:"..."` struct tags for all fields in a struct, so JSON (un)marshaling works.

However, sometimes, you want to omit fields, which can be done with the `x-go-json-ignore` extension.

We can see this at play with the following schemas:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-go-json-ignore
components:
  schemas:
    Client:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        complexField:
          type: object
          properties:
            name:
              type: string
            accountName:
              type: string
          # ...
    ClientWithExtension:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        complexField:
          type: object
          properties:
            name:
              type: string
            accountName:
              type: string
          # ...
          x-go-json-ignore: true
```

From here, we now get two different models:

```go
// Client defines model for Client.
type Client struct {
	ComplexField *struct {
		AccountName *string `json:"accountName,omitempty"`
		Name        *string `json:"name,omitempty"`
	} `json:"complexField,omitempty"`
	Name string `json:"name"`
}

// ClientWithExtension defines model for ClientWithExtension.
type ClientWithExtension struct {
	ComplexField *struct {
		AccountName *string `json:"accountName,omitempty"`
		Name        *string `json:"name,omitempty"`
	} `json:"-"`
	Name string `json:"name"`
}
```

Notice that the `ComplexField` is still generated in full, but the type will then be ignored with JSON marshalling.

You can see this in more detail in [the example code](../examples/extensions/xgojsonignore/).

## `x-oapi-codegen-extra-tags`

Generate arbitrary struct tags to fields.

If you're making use of a field's struct tags to i.e. apply validation, decide whether something should be logged, etc, you can use `x-oapi-codegen-extra-tags` to set additional tags for your generated types.

We can see this at play with the following schemas:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-oapi-codegen-extra-tags
components:
  schemas:
    Client:
      type: object
      required:
        - name
        - id
      properties:
        name:
          type: string
        id:
          type: number
    ClientWithExtension:
      type: object
      required:
        - name
        - id
      properties:
        name:
          type: string
        id:
          type: number
          x-oapi-codegen-extra-tags:
            validate: "required,min=1,max=256"
            safe-to-log: "true"
            gorm: primarykey
```

From here, we now get two different models:

```go
// Client defines model for Client.
type Client struct {
	Id   float32 `json:"id"`
	Name string  `json:"name"`
}

// ClientWithExtension defines model for ClientWithExtension.
type ClientWithExtension struct {
	Id   float32 `gorm:"primarykey" json:"id" safe-to-log:"true" validate:"required,min=1,max=256"`
	Name string  `json:"name"`
}
```

You can see this in more detail in [the example code](../examples/extensions/xoapicodegenextratags/).

## `x-enum-varnames` / `x-enumNames`

Override generated variable names for enum constants.

When consuming an enum value from an external system, the name may not produce a nice variable name. Using the `x-enum-varnames` extension allows overriding the name of the generated variable names.

We can see this at play with the following schemas:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-enumNames and x-enum-varnames
components:
  schemas:
    ClientType:
      type: string
      enum:
        - ACT
        - EXP
    ClientTypeWithNamesExtension:
      type: string
      enum:
        - ACT
        - EXP
      x-enumNames:
        - Active
        - Expired
    ClientTypeWithVarNamesExtension:
      type: string
      enum:
        - ACT
        - EXP
      x-enum-varnames:
        - Active
        - Expired
```

From here, we now get two different forms of the same enum definition.

```go
// Defines values for ClientType.
const (
	ACT ClientType = "ACT"
	EXP ClientType = "EXP"
)

// Defines values for ClientTypeWithNamesExtension.
const (
	ClientTypeWithNamesExtensionActive  ClientTypeWithNamesExtension = "ACT"
	ClientTypeWithNamesExtensionExpired ClientTypeWithNamesExtension = "EXP"
)

// Defines values for ClientTypeWithVarNamesExtension.
const (
	ClientTypeWithVarNamesExtensionActive  ClientTypeWithVarNamesExtension = "ACT"
	ClientTypeWithVarNamesExtensionExpired ClientTypeWithVarNamesExtension = "EXP"
)

// ClientType defines model for ClientType.
type ClientType string

// ClientTypeWithNamesExtension defines model for ClientTypeWithNamesExtension.
type ClientTypeWithNamesExtension string

// ClientTypeWithVarNamesExtension defines model for ClientTypeWithVarNamesExtension.
type ClientTypeWithVarNamesExtension string
```

You can see this in more detail in [the example code](../examples/extensions/xenumnames/).

## `x-deprecated-reason`

Add a GoDoc deprecation warning to a type.

When an OpenAPI type is deprecated, a deprecation warning can be added in the GoDoc using `x-deprecated-reason`.

We can see this at play with the following schemas:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-deprecated-reason
components:
  schemas:
    Client:
      type: object
      required:
        - name
      properties:
        name:
          type: string
        id:
          type: number
    ClientWithExtension:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          deprecated: true
          x-deprecated-reason: Don't use because reasons
        id:
          type: number
          # NOTE that this doesn't generate, as no `deprecated: true` is set
          x-deprecated-reason: NOTE you shouldn't see this, as you've not deprecated this field
```

From here, we now get two different forms of the same enum definition.

```go
// Client defines model for Client.
type Client struct {
	Id   *float32 `json:"id,omitempty"`
	Name string   `json:"name"`
}

// ClientWithExtension defines model for ClientWithExtension.
type ClientWithExtension struct {
	Id *float32 `json:"id,omitempty"`
	// Deprecated: Don't use because reasons
	Name string `json:"name"`
}
```

Notice that because we've not set `deprecated: true` to the `name` field, it doesn't generate a deprecation warning.

You can see this in more detail in [the example code](../examples/extensions/xdeprecatedreason/).

## `x-order`

Explicitly order struct fields.

Whether you like certain fields being ordered before others, or you want to perform more efficient packing of your structs, the `x-order` extension is here for you.

Note that `x-order` is 1-indexed - `x-order: 0` is not a valid value.

We can see this at play with the following schemas:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-order
components:
  schemas:
    Client:
      type: object
      required:
        - name
      properties:
        a_name:
          type: string
        id:
          type: number
    ClientWithExtension:
      type: object
      required:
        - name
      properties:
        a_name:
          type: string
          x-order: 2
        id:
          type: number
          x-order: 1
```

From here, we now get two different forms of the same type definition.

```go
// Client defines model for Client.
type Client struct {
	AName *string  `json:"a_name,omitempty"`
	Id    *float32 `json:"id,omitempty"`
}

// ClientWithExtension defines model for ClientWithExtension.
type ClientWithExtension struct {
	Id    *float32 `json:"id,omitempty"`
	AName *string  `json:"a_name,omitempty"`
}
```

You can see this in more detail in [the example code](../examples/extensions/xorder/).

## `x-oapi-codegen-only-honour-go-name`

Only honour the `x-go-name` when generating field names.

> [!WARNING]
> Using this option may lead to cases where `oapi-codegen`'s rewriting of field names to prevent clashes with other types, or to prevent including characters that may not be valid Go field names.

In some cases, you may not want use the inbuilt options for converting an OpenAPI field name to a Go field name, such as the `name-normalizer: "ToCamelCaseWithInitialisms"`, and instead trust the name that you've defined for the type better.

In this case, you can use `x-oapi-codegen-only-honour-go-name` to enforce this, alongside specifying the `allow-unexported-struct-field-names` compatibility option.

This allows you to take a spec such as:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: x-oapi-codegen-only-honour-go-name
components:
  schemas:
    TypeWithUnexportedField:
      description: A struct will be output where one of the fields is not exported
      properties:
        name:
          type: string
        id:
          type: string
          # NOTE that there is an explicit usage of a lowercase character
          x-go-name: accountIdentifier
          x-oapi-codegen-extra-tags:
            json: "-"
          x-oapi-codegen-only-honour-go-name: true
```

And we'll generate:

```go
// TypeWithUnexportedField A struct will be output where one of the fields is not exported
type TypeWithUnexportedField struct {
	accountIdentifier *string `json:"-"`
	Name              *string `json:"name,omitempty"`
}
```

You can see this in more detail in [the example code](../examples/extensions/xoapicodegenonlyhonourgoname).
