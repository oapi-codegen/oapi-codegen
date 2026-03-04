# Overriding the version of oapi-codegen

## Why?

oapi-codegen uses the standard Go means to determine what the version of the binary is.

However, this doesn't work when there's no Version Control System (VCS), for instance when building from a source bundle.

This example shows how to override the version at build-time.

## How?

By specifying `-ldflags` for the `noVCSVersionOverride` when running `go build` or `go run`:

```sh
go run -ldflags "-X main.noVCSVersionOverride=v123.456.789" ./cmd/oapi-codegen --config=config.yaml ../../api.yaml
```
