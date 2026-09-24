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

## 6. Test case placement in `internal/test/`

`internal/test/` is organized **by feature category**, not by GitHub issue. The top-level categories are:

- `aggregates/` — allOf/anyOf/oneOf composition, anonymous-schema hoisting
- `bodies/` — request/response bodies and content types
- `clients/` — client construction and options
- `events/` — webhooks and callbacks
- `extensions/` — `x-go-*`, `x-oapi-codegen-*`, `x-order`, `x-omitempty` extensions
- `naming/` — identifier generation and type-name collision handling
- `openapi31/` — OpenAPI 3.1-specific behavior
- `options/` — output-options flags (name-normalizer, filters, skip-prune, yaml-tags, …)
- `parameters/` — parameter binding, styles, encoding, nil handling (incl. the cross-framework `roundtrip/` harness)
- `paths/` — path-level routing edge cases: literal colons in paths, URL escaping of reserved characters, and path-parameter precedence. Because colon-in-path routing is backend-specific, leaves here may fan out into per-framework subpackages (`echo/`, `gin/`, `fiber/`, `fiberv3/`, …) each with its own `config.yaml` and generated server, driven by one shared `spec.yaml` and a single top-level routing test — mirroring `parameters/roundtrip/`. This multi-subpackage shape is expected for routing fixtures and is not the "standard leaf layout" churn to flag.
- `references/` — external `$ref`s, import-mapping, multi-package generation, overlays
- `schemas/` — schema-to-type mapping (primitives, objects, enums, nullable, recursive, …)
- `servers/` — server codegen (routers, middleware, strict servers)
- `spec_validation/` — pre-generation spec validation

When a PR adds a test, enforce the following:

- **Do not allow issue-numbered test directories to come back.** Flag any new directory named after a GitHub issue (`internal/test/issues/…`, `issue-1234/`, `issueNNNN/`, and similar) anywhere under `internal/test/`. The old `issues/` tree was deliberately dissolved into the categories above.
- **Prefer extending an existing test case.** If the scenario fits an existing category leaf — same OpenAPI construct, same generation config — the new schemas/operations belong in that leaf's `spec.yaml` and its `*_test.go`, marked with a provenance comment (`# From issue-NNNN: <one-line summary>`) so the issue context is not lost. Suggest the specific leaf to extend when you can identify one.
- **A new leaf is fine when nothing matches.** A new subdirectory under the right category is the correct move when the scenario needs a *different generation config* (other `generate:` targets or `output-options:`) or inherently needs its own files (multi-file external-ref layouts, or per-framework routing fixtures like `paths/` and `parameters/roundtrip/`). It should follow the standard layout — `doc.go` (with the `//go:generate` line), `config.yaml`, `spec.yaml`, `<name>_test.go` — with a snake_case scenario-named directory (never an issue number) and the issue reference in a comment. Cross-framework fixtures legitimately depart from this by fanning out into per-framework subpackages; do not flag that shape.
- **Regression tests for bug fixes are still expected** — they just live in the category matching the feature, not in an issue-named directory.

## 7. General

- The repo is a multi-module monorepo. Cross-module changes (e.g., to `runtime/` consumers) deserve extra scrutiny.
- Generated files are committed; CI fails if `make generate` produces a diff. If a PR's generated files look stale, that will fail CI regardless — but flagging it in review saves a round-trip.

## 5. Schema-merging versions must not regress

`compatibility.schema-merging-behavior` selects how `allOf`, `anyOf` and `oneOf` schemas become Go types, in sequential versions (`v1`, `v2`). These versions still receive bug fixes, but they must never regress: a spec that generated working code under a version must keep generating code that compiles against users' existing code and behaves the same way. New behavior goes into a new version, with code of its own.

The tests that guard these versions are:

- `pkg/codegen/merge_schemas_v2_test.go`, and the tests that cover `v1`: `TestMergeSchemasRecursionUnderOldMergeSchemas` and `TestSchemaMergingV1IsOldMergeSchemas`
- `internal/test/aggregates/matrix/v1/` and `internal/test/aggregates/matrix/v2/`, both the generated code and the tests
- the shape files in `internal/test/aggregates/matrix/shapes/` that run under these versions: those that list `v1` or `v2` under `versions:` (a shape that lists no versions runs under `v2`)
- `internal/test/aggregates/allof/config_old_merge.yaml` and its generated `allof_old_merge.gen.go`, which cover `v1` through `old-merge-schemas`
- other fixtures under `internal/test/` whose config sets `schema-merging-behavior: v2`: they test shapes that `v3` rejects, so they stay on `v2`

**Check that none of these tests is updated to accept a breaking change.** Adding tests is fine, and so is updating them for a bug fix: output that failed to generate, didn't compile, or lost data now works. Changing an existing expectation so that it accepts different Go types (removed or renamed types, fields or methods; changed field types) or different wire behavior for a spec that used to work is a regression: flag it. A commit that changes these tests should say why, so this can be checked. Shape files that these versions use should not be edited to test something new; new cases go into new shape files.

`v3` is the version under development. Its code (`pkg/codegen/merge_schemas_v3.go`, `pkg/codegen/union_v3.go`) and tests (`internal/test/aggregates/matrix/v3/`, `pkg/codegen/*_v3_test.go`) are expected to change as it takes shape; this rule doesn't apply to them.

The code of these versions is `pkg/codegen/merge_schemas_v1.go`, `pkg/codegen/merge_schemas_v2.go` and `pkg/codegen/union_v2.go`, but it also calls code every version shares, such as `generateGoSchema`, `detectEnumViaOneOf`, `schemaIsNullable` and `nonNullTypes` in `pkg/codegen`, and the union and additional-properties templates. A change there can alter these versions' output too, and the tests above are what catch it.
