# Override boilerplate version example
## Why?
oapi-codegen uses default way to define bin's version.

It doesn't work without VCS (git) context.

This example shows how to override version build-time.

## How?
Just add `-ldflags "-X main.noVcsVersionOverride=(YOUR-VERSION)"` to your `go build` or `go run` command:

```
go run -ldflags "-X main.noVcsVersionOverride=overrided" ./cmd/oapi-codegen --config=config.yaml ../../api.yaml
```

Now output DO-NOT-EDIT-guards will contain also non-(devel) version.
 