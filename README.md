OpenAPI Client and Server Code Generator
----------------------------------------

This package contains a set of utilities for generating Go boilerplate code for
services based on
[OpenAPI 3.0](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md)
API definitions. When working with services, it's important to have an API
contract which servers and clients both imprement to minimize the chances of
incompatibilities. It's tedious to generate Go models which precisely correspond to
OpenAPI specifications, so let our code generator do that work for you, so that
you can focus on implementing the business logic for your service.

We have chosen to use [Echo](https://github.com/labstack/echo) as
our default HTTP routing engine, due to its speed and simplicity for the generated
stubs, and [Chi](https://github.com/go-chi/chi) is also supported as an alternative.

This package tries to be too simple rather than too generic, so we've made some
design decisions in favor of simplicity, knowing that we can't generate strongly
typed Go code for all possible OpenAPI Schemas.

## Overview

We're going to use the OpenAPI example of the
[Expanded Petstore](https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore-expanded.yaml)
in the descriptions below, please have a look at it.

In order to create a Go server to serve this exact schema, you would have to
write a lot of boilerplate code to perform all the marshalling and unmarshalling
into objects which match the OpenAPI 3.0 definition. The code generator in this
directory does a lot of that for you. You would run it like so:

    go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
    oapi-codegen petstore-expanded.yaml  > petstore.gen.go

Let's go through that `petstore.gen.go` file to show you everything which was
generated.


## Generated Server Boilerplate

The `/components/schemas` section in OpenAPI defines reusable objects, so Go
types are generated for these. The Pet Store example defines `Error`, `Pet`,
`Pets` and `NewPet`, so we do the same in Go:
```go
// Type definition for component schema "Error"
type Error struct {
    Code    int32  `json:"code"`
    Message string `json:"message"`
}

// Type definition for component schema "NewPet"
type NewPet struct {
    Name string  `json:"name"`
    Tag  *string `json:"tag,omitempty"`
}

// Type definition for component schema "Pet"
type Pet struct {
    // Embedded struct due to allOf(#/components/schemas/NewPet)
    NewPet
    // Embedded fields due to inline allOf schema
    Id int64 `json:"id"`
}

// Type definition for component schema "Pets"
type Pets []Pet
```

It's best to define objects under `/components` field in the schema, since
those will be turned into named Go types. If you use inline types in your
handler definitions, we will generate inline, anonymous Go types, but those
are more tedious to deal with since you will have to redeclare them at every
point of use.

For each element in the `paths` map in OpenAPI, we will generate a Go handler
function in an interface object. Here is the generated Go interface for our
Echo server.

```go
type ServerInterface interface {
    //  (GET /pets)
    FindPets(ctx echo.Context, params FindPetsParams) error
    //  (POST /pets)
    AddPet(ctx echo.Context) error
    //  (DELETE /pets/{id})
    DeletePet(ctx echo.Context, id int64) error
    //  (GET /pets/{id})
    FindPetById(ctx echo.Context, id int64) error
}
```

These are the functions which you will implement yourself in order to create
a server conforming to the API specification. Normally, all the arguments and
parameters are stored on the `echo.Context` in handlers, so we do the tedious
work of of unmarshaling the JSON automatically, simply passing values into
your handlers.

Notice that `FindPetById` takes a parameter `id int64`. All path arguments
will be passed as arguments to your function, since they are mandatory.

Remaining arguments can be passed in headers, query arguments or cookies. Those
will be written to a `params` object. Look at the `FindPets` function above, it
takes as input `FindPetsParams`, which is defined as follows:
 ```go
// Parameters object for FindPets
type FindPetsParams struct {
    Tags  *[]string `json:"tags,omitempty"`
    Limit *int32   `json:"limit,omitempty"`
}
```

The HTTP query parameter `limit` turns into a Go field named `Limit`. It is
passed by pointer, since it is an optional parameter. If the parameter is
specified, the pointer will be non-`nil`, and you can read its value.

If you changed the OpenAPI specification to make the parameter required, the
`FindPetsParams` structure will contain the type by value:
```go
type FindPetsParams struct {
    Tags  *[]string `json:"tags,omitempty"`
    Limit int32     `json:"limit"`
}
```

### Registering handlers
There are a few ways of registering your http handler based on the type of server generated i.e. `-generate server` or `-generate chi-server`

<details><summary><code>Echo</code></summary>

The usage of `Echo` is out of scope of this doc, but once you have an
echo instance, we generate a utility function to help you associate your handlers
with this autogenerated code. For the pet store, it looks like this:
```go
func RegisterHandlers(router codegen.EchoRouter, si ServerInterface) {
    wrapper := ServerInterfaceWrapper{
        Handler: si,
    }
    router.GET("/pets", wrapper.FindPets)
    router.POST("/pets", wrapper.AddPet)
    router.DELETE("/pets/:id", wrapper.DeletePet)
    router.GET("/pets/:id", wrapper.FindPetById)
}
```

The wrapper functions referenced above contain generated code which pulls
parameters off the `Echo` request context, and unmarshals them into Go objects.

You would register the generated handlers as follows:
```go
func SetupHandler() {
    var myApi PetStoreImpl  // This implements the pet store interface
    e := echo.New()
    petstore.RegisterHandlers(e, &myApi)
    ...
}
```

</summary></details>

<details><summary><code>Chi</code></summary>

```go
type PetStoreImpl struct {}
func (*PetStoreImpl) GetPets(r *http.Request, w *http.ResponseWriter) {
    // Implement me
}

func SetupHandler() {
    var myApi PetStoreImpl

    r := chi.Router()
    r.Mount("/", Handler(&myApi))
}
```
</summary></details>

<details><summary><code>net/http</code></summary>

```go
type PetStoreImpl struct {}
func (*PetStoreImpl) GetPets(r *http.Request, w *http.ResponseWriter) {
    // Implement me
}

func SetupHandler() {
    var myApi PetStoreImpl

    http.Handle("/", Handler(&myApi))
}
```
</summary></details>

#### Additional Properties in type definitions

[OpenAPI Schemas](https://swagger.io/specification/#schemaObject) implicitly
accept `additionalProperties`, meaning that any fields provided, but not explicitly
defined via properties on the schema are accepted as input, and propagated. When
unspecified, the `additionalProperties` field is assumed to be `true`.

Additional properties are tricky to support in Go with typing, and require
lots of boilerplate code, so in this library, we assume that `additionalProperties`
defaults to `false` and we don't generate this boilerplate. If you would like
an object to accept `additionalProperties`, specify a schema for `additionalProperties`.

Say we declared `NewPet` above like so:
```yaml
    NewPet:
      required:
        - name
      properties:
        name:
          type: string
        tag:
          type: string
      additionalProperties:
        type: string
```

The Go code for `NewPet` would now look like this:
```go
// NewPet defines model for NewPet.
type NewPet struct {
	Name                 string            `json:"name"`
	Tag                  *string           `json:"tag,omitempty"`
	AdditionalProperties map[string]string `json:"-"`
}
```

The additionalProperties, of type `string` become `map[string]string`, which maps
field names to instances of the `additionalProperties` schema.
```go
// Getter for additional properties for NewPet. Returns the specified
// element and whether it was found
func (a NewPet) Get(fieldName string) (value string, found bool) {...}

// Setter for additional properties for NewPet
func (a *NewPet) Set(fieldName string, value string) {...}

// Override default JSON handling for NewPet to handle additionalProperties
func (a *NewPet) UnmarshalJSON(b []byte) error {...}

// Override default JSON handling for NewPet to handle additionalProperties
func (a NewPet) MarshalJSON() ([]byte, error) {...}w
```

There are many special cases for `additionalProperties`, such as having to
define types for inner fields which themselves support additionalProperties, and
all of them are tested via the `internal/test/components` schemas and tests. Please
look through those tests for more usage examples. 

## Generated Client Boilerplate

Once your server is up and running, you probably want to make requests to it. If
you're going to do those requests from your Go code, we also generate a client
which is conformant with your schema to help in marshaling objects to JSON. It
uses the same types and similar function signatures to your request handlers.

The interface for the pet store looks like this:

```go
// The interface specification for the client above.
type ClientInterface interface {

	// FindPets request
	FindPets(ctx context.Context, params *FindPetsParams) (*http.Response, error)

	// AddPet request with JSON body
	AddPet(ctx context.Context, body NewPet) (*http.Response, error)

	// DeletePet request
	DeletePet(ctx context.Context, id int64) (*http.Response, error)

	// FindPetById request
	FindPetById(ctx context.Context, id int64) (*http.Response, error)
}
```

A Client object which implements the above interface is also generated:

```go
// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
    // The endpoint of the server conforming to this interface, with scheme,
    // https://api.deepmap.com for example.
    Server string

    // HTTP client with any customized settings, such as certificate chains.
    Client http.Client

    // A callback for modifying requests which are generated before sending over
    // the network.
    RequestEditor func(ctx context.Context, req *http.Request) error
}
```

Each operation in your OpenAPI spec will result in a client function which
takes the same arguments. It's difficult to handle any arbitrary body that
Swagger supports, so we've done some special casing for bodies, and you may get
more than one function for an operation with a request body.

1) If you have more than one request body type, meaning more than one media
 type, you will have a generic handler of this form:

        AddPet(ctx context.Context, contentType string, body io.Reader)

2) If you have only a JSON request body, you will get:

        AddPet(ctx context.Context, body NewPet)

3) If you have multiple request body types, which include a JSON type you will
 get two functions. We've chosen to give the JSON version a shorter name, as
 we work with JSON and don't want to wear out our keyboards.

        AddPet(ctx context.Context, body NewPet)
        AddPetWithBody(ctx context.Context, contentType string, body io.Reader)

The Client object above is fairly flexible, since you can pass in your own
`http.Client` and a request editing callback. You can use that callback to add
headers. In our middleware stack, we annotate the context with additional
information such as the request ID and function tracing information, and we
use the callback to propagate that information into the request headers. Still, we
can't foresee all possible usages, so those functions call through to helper 
functions which create requests. In the case of the pet store, we have:

```go
// Request generator for FindPets
func NewFindPetsRequest(server string, params *FindPetsParams) (*http.Request, error) {...}

// Request generator for AddPet with JSON body
func NewAddPetRequest(server string, body NewPet) (*http.Request, error) {...}

// Request generator for AddPet with non-JSON body
func NewAddPetRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {...}

// Request generator for DeletePet
func NewDeletePetRequest(server string, id int64) (*http.Request, error) {...}

// Request generator for FindPetById
func NewFindPetByIdRequest(server string, id int64) (*http.Request, error) {...}
```

You can call these functions to build an `http.Request` from Go objects, which
will correspond to your request schema. They map one-to-one to the functions on
the client, except that we always generate the generic non-JSON body handler.

There are some caveats to using this code.
- exploded, form style query arguments, which are the default argument format
 in OpenAPI 3.0 are undecidable. Say that I have two objects, one composed of
 the fields `(name=bob, id=5)` and another which has `(name=shoe, color=brown)`.
 The first parameter is named `person` and the second is named `item`. The
 default marshaling style for query args would result in
 `/path/?name=bob,id=5&name=shoe,color=brown`. In order to tell what belongs
 to which object, we'd have to look at all the parameters and try to deduce it,
 but we're lazy, so we didn't. Don't use exploded form style arguments if
 you're passing around objects which have similar field names. If you
 used unexploded form parameters, you'd have
 `/path/?person=name,bob,id,5&item=name,shoe,color,brown`, which an be
 parsed unambiguously.

- Parameters can be defined via `schema` or via `content`. Use the `content` form
 for anything other than trivial objects, they can marshal to arbitrary JSON
 structures. When you send them as cookie (`in: cookie`) arguments, we will
 URL encode them, since JSON delimiters aren't allowed in cookies.

## Using SecurityProviders

If you generate client-code, you can use some default-provided security providers
which help you to use the various OpenAPI 3 Authentication mechanism.


```
    import (
        "github.com/deepmap/oapi-codegen/pkg/securityprovider"
    )

    func CreateSampleProviders() error {
        // Example BasicAuth
        // See: https://swagger.io/docs/specification/authentication/basic-authentication/
        basicAuthProvider, basicAuthProviderErr := securityprovider.NewSecurityProviderBasicAuth("MY_USER", "MY_PASS")
        if basicAuthProviderErr != nil {
            panic(basicAuthProviderErr)
        }

        // Example BearerToken
        // See: https://swagger.io/docs/specification/authentication/bearer-authentication/
        bearerTokenProvider, bearerTokenProviderErr := securityprovider.NewSecurityProviderBearerToken("MY_TOKEN")
        if bearerTokenProviderErr != nil {
            panic(bearerTokenProviderErr)
        }

        // Example ApiKey provider
        // See: https://swagger.io/docs/specification/authentication/api-keys/
        apiKeyProvider, apiKeyProviderErr := securityprovider.NewSecurityProviderApiKey("query", "myApiKeyParam", "MY_API_KEY")
        if apiKeyProviderErr != nil {
            panic(apiKeyProviderErr)
        }

        // Example providing your own provider using an anonymous function wrapping in the
        // InterceptoFn adapter. The behaviour between the InterceptorFn and the Interceptor interface
        // are the same as http.HandlerFunc and http.Handler.
        customProvider := func(req *http.Request, ctx context.Context) error {
            // Just log the request header, nothing else.
            log.Println(req.Header)
            return nil
        }

        // Exhaustive list of some defaults you can use to initialize a Client.
        // If you need to override the underlying httpClient, you can use the option
        //
        // WithHTTPClient(httpClient *http.Client)
        //
        client, clientErr := NewClient("https://api.deepmap.com", []ClientOption{
            WithBaseURL("https://api.deepmap.com"),
            WithRequestEditorFn(apiKeyProvider.Edit),
        }...,
        )

        return nil
    }
```

## Using `oapi-codegen`

The default options for `oapi-codegen` will generate everything; client, server,
type definitions and embedded swagger spec, but you can generate subsets of
those via the `-generate` flag. It defaults to `types,client,server,spec`, but
you can specify any combination of those.

- `types`: generate all type definitions for all types in the OpenAPI spec. This
 will be everything under `#components`, as well as request parameter, request
 body, and response type objects.
- `server`: generate the Echo server boilerplate. `server` requires the types in the
 same package to compile.
- `chi-server`: generate the Chi server boilerplate. This code is dependent on
 that produced by the `types` target.
- `client`: generate the client boilerplate. It, too, requires the types to be
 present in its package.
- `spec`: embed the OpenAPI spec into the generated code as a gzipped blob. This
- `skip-fmt`: skip running `go fmt` on the generated code. This is useful for debugging
 the generated file in case the spec contains weird strings.
- `skip-prune`: skip pruning unused components from the spec prior to generating
 the code.

So, for example, if you would like to produce only the server code, you could
run `oapi-generate -generate types,server`. You could generate `types` and
`server` into separate files, but both are required for the server code.

`oapi-codegen` can filter paths base on their tags in the openapi definition.
Use either `-include-tags` or `-exclude-tags` followed by a comma-separated list
of tags. For instance, to generate a server that serves all paths except those
tagged with `auth` or `admin`, use the argument, `-exclude-tags="auth,admin"`.
To generate a server that only handles `admin` paths, use the argument
`-include-tags="admin"`. When neither of these arguments is present, all paths
are generated.

## What's missing or incomplete

This code is still young, and not complete, since we're filling it in as we
need it. We've not yet implemented several things:

- `oneOf`, `anyOf` are not supported with strong Go typing. This schema:

        schema:
          oneOf:
            - $ref: '#/components/schemas/Cat'
            - $ref: '#/components/schemas/Dog'

    will result in a Go type of `interface{}`. It will be up to you
    to validate whether it conforms to `Cat` and/or `Dog`, depending on the
    keyword. It's not clear if we can do anything much better here given the
    limits of Go typing.

    `allOf` is supported, by taking the union of all the fields in all the
    component schemas. This is the most useful of these operations, and is
    commonly used to merge objects with an identifier, as in the
    `petstore-expanded` example.

- `patternProperties` isn't yet supported and will exit with an error. Pattern
 properties were defined in JSONSchema, and the `kin-openapi` Swagger object
 knows how to parse them, but they're not part of OpenAPI 3.0, so we've left
 them out, as support is very complicated.


## Making changes to code generation

The code generator uses a tool to inline all the template definitions into
code, so that we don't have to deal with the location of the template files.
When you update any of the files under the `templates/` directory, you will
need to regenerate the template inlines:

    go generate ./pkg/codegen/templates

All this command does is inline the files ending in `.tmpl` into the specified
Go file.

Afterwards you should run `go generate ./...`, and the templates will be updated
 accordingly.

Alternatively, you can provide custom templates to override built-in ones using
the `-templates` flag specifying a path to a directory containing templates
files. These files **must** be named identically to built-in template files
(see `pkg/codegen/templates/*.tmpl` in the source code), and will be interpreted
on-the-fly at run time. Example:

    $ ls -1 my-templates/
    client.tmpl
    typedef.tmpl
    $ oapi-codegen \
        -templates my-templates/ \
        -generate types,client \
        petstore-expanded.yaml
