# Test category scaffold (`chore/test-cleanup`)

This directory is the **proposed re-organization** of `internal/test/` into a small set
of feature categories. Each leaf is one focused test unit: a shared `spec.yaml`, a
`config.yaml`, and a `doc.go` (`//go:generate`). Today's `internal/test/issues/` and the
per-framework duplication collapse into these categories.

## Why `_categories/` (and not the real layout yet)

The Go toolchain **ignores directories whose names begin with `_`**. Staging here means
these stub packages and dormant `//go:generate` directives cannot affect `make test`,
`make generate`, or lint, and they don't collide with the existing `parameters/`,
`schemas/`, `extensions/`, etc. directories. This is a **scaffold for inspection**.
Promotion = migrate spec/test content in, then move each leaf up to `internal/test/<category>/…`
and delete the old dirs.

## Conventions

- **One spec per leaf.** Each folded case is a named schema/operation in that leaf's
  `spec.yaml`, tagged with a provenance comment: `# issue #2031: nil optional array must be omitted, not null`.
- **Type-shape tests are framework-agnostic** (`models: true`). Only behavior tests
  (`servers/`, `events/`, `parameters/roundtrip`) fan out across HTTP frameworks.
- **Shared models via import-mapping.** `references/multipackage` defines the canonical
  models; downstream categories import-map them instead of redeclaring (the `externalref/` pattern).

## Top-level categories

| Category | What it covers |
|---|---|
| `schemas/` | type generation — primitives, objects, nullable, enums, recursive, readonly/writeonly, defaults, deprecated, security |
| `aggregates/` | allOf / anyOf / oneOf / anonymous-hoisting (reimplementation pending) |
| `openapi31/` | 3.1 idiom **detection** — enum-via-oneOf, content keywords, const/examples polish |
| `parameters/` | path/query/header/cookie/content/encoding/precedence + one multi-framework roundtrip harness |
| `bodies/` | request/response bodies & content-type handling |
| `servers/` | router codegen, strict servers, per-operation middleware (multi-framework) |
| `clients/` | client codegen |
| `events/` | webhooks (3.1) & callbacks (3.0) |
| `references/` | external $ref, multi-package/import-mapping, overlays |
| `extensions/` | x-go-type, skip-optional-pointer, struct-tags, x-order, x-go-name, x-deprecated-reason |
| `options/` | name-normalizer, type-aliases, yaml-tags, filter, skip-prune, compatibility, response-getters |
| `naming/` | conflict resolution & identifier normalization |

The migration mapping (every current dir + all 56 `issues/*` → destination leaf) lives in
the PR description / planning notes; each leaf's `spec.yaml` header repeats the subset that
folds into it.
