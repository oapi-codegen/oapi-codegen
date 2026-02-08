# `oapi-codegen` V3

This is an experimental prototype of a V3 version of oapi-codegen. The generated code and command line options are not yet stable. Use at your
own risk.

## What is new in Version 3
This directory contains an experimental version of oapi-codegen's future V3 version, which is based on [libopenapi](https://github.com/pb33f/libopenapi),
instead of the prior [kin-openapi](https://github.com/getkin/kin-openapi). This change necessitated a nearly complete rewrite, but we strive to be as
compatible as possible.

What is working:
  - All model, client and server generation as in earlier versions.
  - We have added Webhook and Callback support, please see `./examples`, which contains the ubiquitous OpenAPI pet shop implemented in all supported servers
    and examples of webhooks and callbacks implemented on top of the `http.ServeMux` server, with no additional imports.
  - Model generation of `allOf`, `anyOf`, `oneOf` is much more robust, since these were a source of many problems in earlier versions.
  - Echo V5 support has been added (Go 1.25 required)

What is missing:
  - We have not yet created any runtime code like request validation middleware.

## Differences in V3

V3 is a brand new implementation, and may (will) contain new bugs, but also strives to fix many current, existing bugs. We've run quite a few
conformance tests against specifications in old Issues, and we're looking pretty good! Please try this out, and if it failes in some way, please
file Issues.

### Aggregate Schemas

V3 implements `oneOf`, `anyOf`, `allOf`  differently. Our prior versions had pretty good handling for `allOf`, where we merge all the constituent schemas
into a schema object that contains all the fields of the originals. It makes sense, since this is composition.

`oneOf` and `anyOf` were handled by deferred parsing, where the JSON was captured into `json.RawMessage` and helper functions were created on each
type to parse the JSON as any constituent type with the user's help.

Given the following schemas:

```yaml
components:
  schemas:
    Cat:
      type: object
      properties:
        name:
          type: string
        color:
          type: string
    Dog:
      type: object
      properties:
        name:
          type: string
        color:
          type: string
    Fish:
      type: object
      properties:
        species:
          type: string
        isSaltwater:
          type: boolean
    NamedPet:
      anyOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'
    SpecificPet:
      oneOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'
        - $ref: '#/components/schemas/Fish'
```

#### V2 output

V2 generates both `NamedPet` (anyOf) and `SpecificPet` (oneOf) identically as opaque wrappers around `json.RawMessage`:

```go
type NamedPet struct {
	union json.RawMessage
}

type SpecificPet struct {
	union json.RawMessage
}
```

The actual variant types are invisible at the struct level. To access the underlying data, the user must call generated helper methods:

```go
// NamedPet (anyOf) helpers
func (t NamedPet) AsCat() (Cat, error)
func (t *NamedPet) FromCat(v Cat) error
// <snip>

// SpecificPet (oneOf) helpers

func (t SpecificPet) AsFish() (Fish, error)
func (t *SpecificPet) FromFish(v Fish) error
func (t *SpecificPet) MergeFish(v Fish) error
// <snip>
```

Note that `anyOf` and `oneOf` produce identical types and method signatures; there is no semantic difference in the generated code.

#### V3 output

V3 generates structs with exported pointer fields for each variant, making the union members visible at the type level. Crucially, `anyOf` and `oneOf` now have different marshal/unmarshal semantics.

**`anyOf` (NamedPet)** — `MarshalJSON` merges all non-nil variants into a single JSON object. `UnmarshalJSON` tries every variant and keeps whichever succeed:

```go
type NamedPet struct {
	Cat *Cat
	Dog *Dog
}

```

**`oneOf` (SpecificPet)** — `MarshalJSON` returns an error unless exactly one field is non-nil. `UnmarshalJSON` returns an error unless the JSON matches exactly one variant:

```go
type SpecificPet struct {
	Cat  *Cat
	Dog  *Dog
	Fish *Fish
}
```

### OpenAPI V3.1 Feature Support

Thanks to [libopenapi](https://github.com/pb33f/libopenapi), we are able to parse OpenAPI 3.1 and 3.2 specifications. They are functionally similar, you can
read the differences between `nullable` fields yourself, but they add some new functionality, namely `webhooks` and `callbacks`. We support all of them in
this prototype. `callbacks` and `webhooks` are basically the inverse of `paths`. Webhooks contain no URL element in their definition, so we can't register handlers
for you in your http router of choice, you have to do that yourself. Callbacks support complex request URL's which may reference the original request. This is
something you need to pull out of the request body, and doing it generically is difficult, so we punt this problem, for now, to our users.

Please see the [webhook example](examples/webhook/). It creates a little server that pretends to be a door badge reader, and it generates an event stream
about people coming and going. Any number of clients may subscribe to this event. See the [doc.go](examples/webhook/doc.go) for usage examples.

The [callback example](examples/callback), creates a little server that pretends to plant trees. Each tree planting request contains a callback to be notified
when tree planting is complete. We invoke those in a random order via delays, and the client prints out callbacks as they happen. Please see [doc.go](examples/callback/doc.go) for usage.

### Flexible Configuration

oapi-codegen V3 tries to make no assumptions about which initialisms, struct tags, or name mangling that is correct for you. A very [flexible configuration file](Configuration.md) allows you to override anything.


### No runtime dependency

V2 generated code relied on `github.com/oapi-codegen/runtime` for parameter binding and styling. This was a complaint from lots of people due to various
audit requirements. V3 embeds all necessary helper functions and helper types into the spec. There are no longer generic, parameterized functions that
handle arbitrary parameters, but rather very specific functions for each kind of parameter, and we call the correct little helper versus a generic
runtime helper.

### Models now support default values configured in the spec

Every model which we generate supports an `ApplyDefaults()` function. It recursively applies defaults on
any unset optional fields. There's a little caveat here, in that some types are external references, so
we call `ApplyDefaults()` on them via reflection. This might call an `ApplyDefaults()` which is completely
unrelated to what we're doing. Please let me know if this feature is causing trouble.

## Installation

Go 1.24 is required, install like so:

    go get -tool github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen@latest

You can then run the code generator

    //go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen

