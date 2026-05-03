# Code Review Rules for oapi-codegen

These rules guide automated code review for this repository. oapi-codegen is a code generator that turns OpenAPI 3.x specs into Go server stubs, clients, and types. The primary risk surface in any change is **the generated output that downstream users compile against** — review with that in mind.

## 1. Generated files (`*.gen.go`)

Files matching `*.gen.go` are produced by the code generator and committed to source control so CI can verify they are up-to-date. You do **not** need to read every generated file in a PR — spot check a representative sample.

When spot checking, look for:

- **Drift unrelated to the stated change.** If the PR claims to fix X, but a `*.gen.go` diff also shows unrelated reshuffling (renamed identifiers, reordered methods, formatting churn, removed comments, regenerated imports), call it out. This often indicates an accidental commit from a different branch, an out-of-date local toolchain, or a template change the author did not intend.
- **Diffs that look "too small" for the described change.** If the PR claims to add a new code generation feature but the only generated diff is a one-line tweak in one file, the test fixtures may not have been regenerated — flag it.
- **Diffs that look "too large" for the described change.** A small template tweak should not produce thousands of lines of churn across every fixture. If it does, the template change probably has a wider blast radius than the author realized.

You do not need to suggest stylistic improvements inside `*.gen.go` files — they are templated output, not hand-written code.

## 2. Configuration changes must update the JSON schema

Whenever `pkg/codegen/configuration.go` is modified — particularly the `Configuration`, `GenerateOptions`, `OutputOptions`, or `CompatibilityOptions` structs — the corresponding JSON Schema at `configuration-schema.json` (repo root) **must** be updated to match.

Flag any PR that changes `pkg/codegen/configuration.go` without a corresponding edit to `configuration-schema.json`. The schema is consumed by IDEs and validation tooling; an out-of-sync schema silently breaks downstream users' configs.

## 3. Watch for breaking changes to the generated API surface

Generated code is compiled into downstream users' binaries. Any change that alters the **shape** of generated code is a potential breaking change for every user of oapi-codegen, even if the change is "just" in a template.

Flag the following patterns and call them out as breaking-change risks:

- **Function/method signature changes** in generated server interfaces, client methods, or strict server handler types — added/removed/reordered parameters, changed return types, changed receiver types.
- **Removed or renamed exported types, functions, methods, fields, or constants** in generated output.
- **Changed Go types for existing fields** (e.g., `*string` → `string`, `int` → `int64`, switching between value and pointer receivers, swapping a concrete type for an interface).
- **Renamed JSON struct tags** or other tags that affect serialization wire format.
- **Reordered struct fields** when the struct is embedded or used in positional contexts (rare but worth noting).
- **Removed or renamed template helper functions** that user-supplied template overrides may depend on (templates in `pkg/codegen/templates/` and helpers in `pkg/codegen/template_helpers.go`).

Per the project's stated practice, behavior-changing features in codegen should be **opt-in via configuration** (new flags under `generate`, `output-options`, or `compatibility`), not silent breaking changes. If you see a behavior change that is unconditional, ask whether it should be gated behind a new compatibility flag.

## 4. Template changes should be evaluated across all router backends

The repo supports multiple server frameworks via separate template subdirectories under `pkg/codegen/templates/`:

- `chi/`
- `echo/`
- `fiber/`
- `gin/`
- `gorilla/`
- `iris/`
- `stdhttp/` (net/http ServeMux, Go 1.22+)
- `strict/` (the strict-server wrapper layer, applied on top of any backend)

Plus shared templates at the top level: `client.tmpl`, `client-with-responses.tmpl`, `typedef.tmpl`, `param-types.tmpl`, `request-bodies.tmpl`, `inline.tmpl`, `imports.tmpl`, `constants.tmpl`, `server-urls.tmpl`, `additional-properties.tmpl`, `union.tmpl`, `union-and-additional-properties.tmpl`.

When a PR modifies a template, ask:

- **Does this change apply to other backends?** A bug fix or new feature in one backend's template often needs an analogous fix in the others. The implementation will not be identical — each framework has its own routing, middleware, and parameter-binding idioms — but the *intent* should usually be applied everywhere it is relevant.
- **If only one backend is touched, is that intentional?** A change scoped to a single backend may be correct (e.g., a Fiber-specific bug, a Gin-specific middleware quirk). If so, the PR description should explain why. If there is no such explanation and the change looks generally applicable, flag it.
- **Did the strict-server templates need a corresponding update?** Strict-mode wraps the per-backend handlers and frequently needs parallel changes.
- **Were the integration tests in `internal/test/` updated?** This module imports every framework and is the primary place backend-parity bugs surface.

Be pragmatic: do not demand identical changes in seven places. Do your best to assess whether a change is conceptually backend-agnostic (most template changes are) or genuinely backend-specific (some are), and flag missing parity only when the change looks generally applicable.

## 5. Dependencies

- Watch out for users making `go.mod` changes which advance the Go version to a new minor release. We want these to be explicitly justified in the commit message. Maintenance version bumps are ok. Mention that the maintainers would prefer to do a minor version bump themselves.
- Watch out for unnecessary dependency changes that creep into code reviews, just because people tend to do it out of habit. Every dependency update should have a reason if it's bundled with codegen changes.

## 6. General

- The repo is a multi-module monorepo. Cross-module changes (e.g., to `runtime/` consumers) deserve extra scrutiny.
- `internal/test/` contains regression tests keyed to GitHub issues. New bug fixes should generally include a regression test there.
- Generated files are committed; CI fails if `make generate` produces a diff. If a PR's generated files look stale, that will fail CI regardless — but flagging it in review saves a round-trip.
