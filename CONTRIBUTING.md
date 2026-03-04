## Contributing

If you're interested in contributing to `oapi-codegen`, the first thing we have to say is thank you! We'd like to extend our gratitude to anyone who takes the time to improve this project.

`oapi-codegen` is being actively maintained, however the two people who do so are very busy, and can only set aside time for this project every once in a while, so our release cadence is slow and conservative.

> [!NOTE]
> We're actively considering what needs to change to make `oapi-codegen` more sustainable, and hope that we can share soon some options.

This guide is a starting point, and we'll absolutely improve it on an ongoing basis. We've managed to go ~4 years without a substantial guide like this - sometimes to the detriment of contributors - and would love to keep improving this guide, and the project, for the best of the community.

### When may we not change things?

Generating code which others depend on, which is based on something as complex as OpenAPI is fraught with many edge cases, and we prefer to leave things as they are if there is a reasonable workaround.

We'll try to avoid adding too much noise into generated code, or introduce breaking changes (as per SemVer). See also "Backwards compatibility" in the README.

### Raising a bug

If you believe you have encountered a bug, please raise an issue.

> [!TIP]
> Please follow the [minimal reproductions](#minimal-reproductions) documentation to improve our ability to support triaging

This may get converted into a feature request if we don't deem it a bug, but a missing feature.

### Asking a question

We'd prefer that questions about "how do I use (this feature)?" or "what do the community think about ...?" get asked using [GitHub Discussions](https://github.com/oapi-codegen/oapi-codegen/discussions) which allow the community to answer them more easily.

### Making changes that tweak generated code

If you are making changes to the codebase that affects the code that gets generated, you will need to make sure that you have regenerated any generated test cases in the codebase using `make generate`.

These generated test cases and examples provide a means to not only validate the functionality, but as they are checked in to source code, allow us to see if there are any subtle issues or breaking changes.

> [!NOTE]
> Significant changes to generated code are unlikely to be merged, especially in cases where there would be a breaking change that all consumers would have to respond to i.e. renaming a function or changing the function signature.
>
> However, if we can make this an opt-in feature (using the `output-options` configuration object) then that would be our preference.

### Feature enhancements

It's great that you would like to improve `oapi-codegen` and add new futures.

We would prefer there be an issue raised for a feature request first, especially if it may be a duplicate of existing requests. However, sometimes that isn't possible - or takes longer than the code changes required - so it can be excused.

Features that amend the way existing codegen works should - ideally - be behind an opt-in feature flag using the `output-options` configuration object.

### Minimal reproductions

> [!TIP]
> The minimal reproductions for bugs may get taken into the codebase (licensed under `Apache-2.0`) as a test-case for future regression testing
>
> However, this can only be done if you license the code under `Apache-2.0` itself - if you are comfortable doing so, please do.

When raising a bug report, or asking a question about functionality, it's super helpful if you can share:

- The version of `oapi-codegen` you're using
  - You _may_ get asked to update to a later - or latest - version, to see if the issue persists
- The YAML configuration file you're using
- The OpenAPI spec you're using
  - However, we would prefer it only be the _absolute minimum_ specification, to limit the noise while trying to debug the issue, and to reduce information exposure from internal API development
- What problem you're seeing
- What the expected behaviour is
- What version of Go you're using

> [!CAUTION]
> When sharing a minimal reproduction, please be aware of sharing any internal information about the APIs you're developing, or any sensitive Intellectual Property.

### Before you raise a PR

> [!NOTE]
> Please raise PRs from a branch that isn't the `master` or `main` branch on your repo. This generally means that as maintainers, we can't push changes to the branch directly.

Before you send the PR, please run the following commands locally:

```sh
make tidy
make test
make generate
make lint
```

It is important to use the `make` tasks due to the way we're (ab)using the Go module system to split the project into multiple modules to reduce our dependency bloat in the main module.

These are also run in GitHub Actions, across a number of Go releases.

It's recommended to raise a draft PR first, so you can get feedback on the PR from GitHub, and review your own changes, before getting the attention of a maintainer.

### "Should I @-mention the maintainers on an issue"

Please try to avoid pinging the maintainers in an issue, Pull Request, or discussion.

> [!NOTE]
> We're actively considering what needs to change to make `oapi-codegen` more sustainable, and hope that we can share soon some options.

The project is run on a volunteer basis, and as such, tagging us on issues - especially if you've just raised them - is largely unhelpful. We monitor the issues and work to triage them as best we can with the time we have allocated for it.
