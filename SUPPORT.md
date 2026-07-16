# Support model

[`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) is currently supported in a best-efforts means, due to the [Core Maintainers](https://github.com/oapi-codegen/governance/#core-maintainer) working in their "off hours" from their busy jobs.

We do thoroughly appreciate our users, the feature requests and bug reports raised, and want to set expectations accordingly.

Related:

- [Creating a more sustainable model for `oapi-codegen` in the future](https://github.com/oapi-codegen/oapi-codegen/discussions/1606)
- [Looking back at `oapi-codegen`'s last year](https://github.com/oapi-codegen/oapi-codegen/discussions/1985)

## Supported versions

Only the latest minor release version of `oapi-codegen` is supported for active development.

`oapi-codegen` does not currently backport any bug fixes.

## Security updates

Related: [`oapi-codegen`'s organisational security policy (`SECURITY.md`)](https://github.com/oapi-codegen/.github/blob/HEAD/SECURITY.md).

## Minimum required Go toolchain version

As per [the install instructions](https://github.com/oapi-codegen/oapi-codegen/#install), `oapi-codegen`'s recommended installation model is as a source-tracked dependency, for instance using `go tool`.

Because of this, we take more care to resist bumping the `go` directive, as it has a knock-on effect for all consumers of `oapi-codegen`, as it requires the consumer to _also_ bump their `go` directive.

When considering whether to bump the `go` directive, we will consider:

- Do we _definitely_ need to pull in this new version of Go?
  - Can we work around it by not using new language features?
- If this is a requirement of an upstream dependency, can upstream use build tags ([like so](https://github.com/charmbracelet/log/pull/13)), to allow us to continue using old versions of Go?
  - If it is a requirement, and we don't want to bump it (yet), can we avoid the Go version bump?
- Is the new version supported by the Go team?
  - We're comfortable not requiring the Go version being in active support - it's up to consumers to decide what toolchain + standard library version they want to use to build

We will not mandate a `toolchain` directive.

## Additional support

For additional support, it's worth reading [oapi-codegen/governance: Sponsorship](https://github.com/oapi-codegen/governance/#sponsorship), and visiting the different [funding options](https://github.com/oapi-codegen/oapi-codegen/blob/main/.github/FUNDING.yml).
