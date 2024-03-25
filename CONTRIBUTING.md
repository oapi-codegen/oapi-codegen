## Contributing

If you're interested in contributing to `oapi-codegen`, the first thing we have to say is thank you! We'd like to extend our gratitude to anyone who takes the time to improve this project.

`oapi-codegen` is being actively maintained, however the two people who do so are very busy, and can only set aside time for this project every once in a while, so our release cadence is slow and conservative.

> [!NOTE]
> We're actively considering what needs to change to make `oapi-codegen` more sustainable, and hope that we can share soon some options.


```
Generating code which others depend on, which is based on something as complex
as OpenAPI is fraught with many edge cases, and we prefer to leave things as
they are if there is a reasonable workaround.

If you do find a case where oapi-codegen is broken, and would like to submit a PR,
we are very grateful, and will happily look at it.

Since most commits affect generated code, before sending your PR, please
ensure that all boilerplate has been regenerated. You can do this from the top level
of the repository by running:

    make generate

I realize that our code isn't entirely idiomatic with respect to comments, and
variable naming and initialisms, especially the generated code, but I'm reluctant
to merge PR's which change this, due to the breakage they will cause for others. If
you rename anything under `/pkg` or change the names of variables in generated
code, you will break other people's code. It's safe to rename internal names.


```

This guide is a starting point, and we'd absolutely **??**. We've managed to go ~4 years without a substantial guide like this, and would love to keep improving it for the best of the community.

### Raising a bug

If you believe **??**.

This may get converted into a feature request if needed.

### Asking a question

We'd prefer that questions about "how do I use (this feature)" or **??** get asked using [GitHub Discussions](https://github.com/deepmap/oapi-codegen/discussions) which allow **??**

### Making changes that tweak generated code

If you are generating **??**.

These provide a **??**, and as they are checked in to source code, allow us to **??**.

Now, please note that significant changes to generated code are likley to **??**, especially in cases where there would be a breaking change issued to **??**.

### Feature enhancements

**Should introduce a flag if possible**

**Opt-in by default?**


### Minimal reproductions

### Before you raise a PR

Before you send the PR, please run the following commands locally:

```sh
make tidy
make test
make generate
make lint
```

It is important to use the `make` tasks due to the way we're (ab)using the Go module system to make our **??**.

These are also run in GitHub Actions, across a number of Go releases.

It's recommended to raise a draft PR first, so you can get
