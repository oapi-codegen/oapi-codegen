# Authenticated API Example

This little server is an example demonstrating how JWT's can be handled somewhat
automatically with per-path validation of scopes.

We use similar code in production at DeepMap, and this example shows how authenticated
API's can be written, as well as unit tested.

Some parts of this code can be generalized in the future, but for now, feel free
to copy/paste and modify this example code.

## API Overview

We define a trivial server which allows for creating and listing `Thing`s, please
see `api.yaml`.

In `#components/securitySchemes` we define a scheme named `BearerAuth`:

```yaml
BearerAuth:
  type: http
  scheme: bearer
  bearerFormat: JWT
```

This security scheme is a global requirement for all API endpoints:

```yaml
security:
  - BearerAuth: [ ]

```

This means that all API endpoints require a JWT bearer token for access, and
without any specific scopes, as denoted by `[]`.

However, we want our `addThing` operation to require a special write permission,
denoted by `things:w`. This is a convention that we use in naming scopes, 
noun followed by type of access, in this case, `:w` means write. Read permission
is implicit from having a valid JWT.

To specialize the `addThing` handler, we override the global security with local
security on `POST /things`:
```yaml
security:
  - BearerAuth:
    - "things:w"
```

## Implementation

Security is tricky, so we need to leverage well tested libraries for doing validation
instead of implementing too much ourselves. We've chosen to use the excellent
[github.com/lestrrat-go/jwx](https://github.com/lestrrat-go/jwx) library for JWT
validation, and the [kin-openapi](https://github.com/getkin/kin-openapi/tree/master/openapi3filter)
request filter to help us perform validation.

First, we need to configure our [OapiRequestValidator](https://github.com/deepmap/oapi-codegen/blob/master/pkg/middleware/oapi_validate.go)
to perform authentication:
```go
validator := middleware.OapiRequestValidatorWithOptions(spec,
    &middleware.Options{
        Options: openapi3filter.Options{
            AuthenticationFunc: NewAuthenticator(v),
        },
    })
```

Whenever a request comes in, `openapi3filter` will call the `AuthenticationFunc` to tell
it whether the request is valid. Please see the `Authenticate` function in
`server/jwt_authenticator.go` for an example how to do this.

In this example, we set up several components:
1) We create a `FakeAuthenticator`. This is a little helper which uses an ECDSA key
 to sign JWT's via the `lestrrat-go/jwx` tools. In a normal application deployment,
 you would be using an identity provider, say, Google, Auth0, AWS Cognito, etc, to
 give you JWT's via some auth protocol, but we wanted to keep it simple. You would
 _never_ have key material present like this inside your code in a real application.
2) The JWT validation part in our fake authenticator is reasonably thorough, and can
 be used as an example for production code. When an authorization header comes in,
 bearing an encoded JWT Token (called JWS), we validate that it was signed by
 the public key of our IDP. In the real world, this would be a service, in this
 example, this is our `FakeAuthenticator`
3) Once the JWT is a legitimate one, signed by the expected authority, we start looking
 at claims. JWT's are very freeform, so you can put in whatever payload you like, since
 the implementation is up to you. As you see in the spec section above, we created
 a the scope `things:w` to denote permission to write `Thing`s. What we expect is that
 the JWT contains a claim named `perms`, that's an array of permissions granted, and
 that it contains the string `things:w`. The `Authenticate` function does this
 check.

## Example

You can run the example Echo server like so:
```
$ go run ./examples/authenticated-api/echo/main.go
2021/10/07 14:32:45 Reader token eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.Hf9dCNJLa0HQfbtJi7ndASbkTfrLc6bZBJK8HaPqtzXiDkTH6sMRoiNhf6Kb1g6z3R1tN3XEpXsghxlMRO3OLA
2021/10/07 14:32:45 Writer token eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOlsidGhpbmdzOnciXX0.CbPT1hzWmyTt0lTyv-fiyUlnY1SGa0vrX52yFjeigx2PA1-78LVH0z5hukPKkLMPDMXL9AJrtNp0elWSD_qrBw

   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v4.2.1
High performance, minimalist Go web framework
https://echo.labstack.com
____________________________________O/_______
                                    O\
```

It prints out two tokens for you to use in http requests. Let's assign
these to some environment variables for convenience.
```shell
export RJWT=eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.Hf9dCNJLa0HQfbtJi7ndASbkTfrLc6bZBJK8HaPqtzXiDkTH6sMRoiNhf6Kb1g6z3R1tN3XEpXsghxlMRO3OLA
export WJWT=eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOlsidGhpbmdzOnciXX0.CbPT1hzWmyTt0lTyv-fiyUlnY1SGa0vrX52yFjeigx2PA1-78LVH0z5hukPKkLMPDMXL9AJrtNp0elWSD_qrBw
```

Let's see how this works in practice. My example commands are using `HTTPie`, which
I find easier to use in a shell than `curl`:

Unauthenticated requests fail:
```
$ http http://localhost:8080/things
HTTP/1.1 403 Forbidden
Content-Length: 43
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Oct 2021 22:00:32 GMT

{
    "message": "Security requirements failed"
}

$ http POST http://localhost:8080/things name=SomeThing
HTTP/1.1 403 Forbidden
Content-Length: 43
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Oct 2021 22:01:11 GMT

{
    "message": "Security requirements failed"
}
```

Using the Writer JWT, we can insert a `Thing` into the server:
```
$ http POST http://localhost:8080/things name=SomeThing Authorization:"Bearer $WJWT"
HTTP/1.1 201 Created
Content-Length: 28
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Oct 2021 22:02:05 GMT

{
    "id": 0,
    "name": "SomeThing"
}
```

However, we can not insert a `Thing` using the reader JWT:
```
$ http POST http://localhost:8080/things name=SomeThing2 Authorization:"Bearer $RJWT"
HTTP/1.1 403 Forbidden
Content-Length: 43
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Oct 2021 22:02:39 GMT

{
    "message": "Security requirements failed"
}
```

Both JWT's, however, permit listing `Things`:
```
$ http http://localhost:8080/things Authorization:"Bearer $RJWT"
HTTP/1.1 200 OK
Content-Length: 30
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Oct 2021 22:03:12 GMT

[
    {
        "id": 0,
        "name": "SomeThing"
    }
]

$ http http://localhost:8080/things Authorization:"Bearer $WJWT"
HTTP/1.1 200 OK
Content-Length: 30
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Oct 2021 22:03:34 GMT

[
    {
        "id": 0,
        "name": "SomeThing"
    }
]
```

