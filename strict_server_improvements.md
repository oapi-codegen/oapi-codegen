# Strict-server template cleanup ideas

The strict-server templates
(`pkg/codegen/templates/strict/strict-interface.tmpl`,
`strict-fiber-interface.tmpl`, `strict-iris-interface.tmpl`,
`strict-responses.tmpl`) carry a lot of decision-making in inline
template conditionals. As the type-emission rules grow we should
keep elevating those decisions into Go and exposing the result as a
small set of named fields/methods on `ResponseContentDefinition`.
The template then keys off those names instead of re-deriving the
same conditions.

The two follow-ups below were called out during the
"Overhaul anonymous schema hoisting" review and deferred. Pick them
up when next touching this code.

## Background: a path we considered and abandoned

A `SkipTypeDeclaration bool` field on `ResponseContentDefinition`
was prototyped to suppress the strict envelope's type declaration
when the underlying schema's `RefType` matched the receiver name —
i.e., when the response-root hoist already declared the type and
the strict template would otherwise emit a self-referential alias
(`type X = X`). It worked for the no-headers alias branch but
*didn't* fix the with-headers struct branch (lines 62–80 of
`strict-interface.tmpl`), where `Body <TypeDecl>` would still
recursively reference the receiver name.

The cleaner fix that landed instead: have the response-root hoist
use a `…ResponseBody`-suffixed name distinct from the strict
envelope, so the envelope can reference the body as
`Body <Op><Status><Tag>ResponseBody` without colliding. With the
names disjoint, no template-level suppression is needed.
`SkipTypeDeclaration` was removed before merge.

The lesson generalizes: prefer making the names disjoint at the
*Go-side* schema construction so the templates stay
suppression-free.

## 1. `EncodeAccessor` (or method) on `ResponseContentDefinition`

**Today** the JSON-encode line in each strict template inlines a
two-axis decision:

```gotmpl
Encode(response{{if $hasBodyVar}}.Body{{end}}{{if and $hasUnionElements (not .Schema.IsExternalRef)}}.union{{end}})
```

The `.union` accessor reaches an unexported field directly, which
only works inside the declaring package. The
`(not .Schema.IsExternalRef)` guard exists solely to keep the
access in-package. That's a Go correctness concern, not a template
concern.

**Suggested** lift the trailing accessor to a method:

```go
// EncodeAccessor returns the trailing accessor on `response` for
// JSON encoding. Returns ".union" only when the union RawMessage
// is reachable in the local package; otherwise "" so the encode
// goes through the type's MarshalJSON.
func (r ResponseContentDefinition) EncodeAccessor() string {
    if len(r.Schema.UnionElements) != 0 && !r.Schema.IsExternalRef() {
        return ".union"
    }
    return ""
}
```

Templates simplify to:

```gotmpl
Encode(response{{if $hasBodyVar}}.Body{{end}}{{.EncodeAccessor}})
```

`$hasBodyVar` itself is a candidate for the same treatment (it's
computed in the template as `or ($hasHeaders) (not $fixedStatusCode)
(not .IsSupported)`).

## 2. Precompute the entire strict envelope declaration line

**Today** lines 41–88 of `strict-interface.tmpl` (and the
fiber/iris variants) chain through five mutually-exclusive
declaration forms (`multipart-func` / `alias` / `defined` /
`inline-struct` / `io-reader`), and immediately follow with an
optional MarshalJSON/UnmarshalJSON pair gated on
`(.IsJSON && .Schema.HasCustomMarshalJSON)`.

**Suggested** move the form-selection into a Go helper that returns
the full declaration block as a string:

```go
func (r ResponseContentDefinition) StrictTypeDeclaration(opID, statusCode string) string {
    // returns "type X = Y", or "type X Y", or "type X func(...)",
    // or the full multi-line block including
    // MarshalJSON/UnmarshalJSON pair.
}
```

Templates collapse to `{{.StrictTypeDeclaration $opid $statusCode}}`.

**Tradeoffs:**
- Loses some template-side scannability — the variants are no
  longer visible as a chain of `if`/`else` blocks.
- The `$ref`-based block at lines 42–48 (when the response is
  itself a `$ref`) uses different inputs (`$ref`, `$isExternalRef`)
  and would need its own helper, which dilutes the simplification.
- Best done together with extracting `$ref` / `$isExternalRef` /
  `$fixedStatusCode` / `$hasHeaders` / `$hasBodyVar` into Go-side
  precomputed fields on `ResponseContentDefinition` or
  `ResponseDefinition`.

This is the bigger refactor and only worth it if the templates
keep growing or we add another envelope form.

## Why this list, in this order

`EncodeAccessor` is high-value, low-cost: the `.union`
in-package-only nuance is a Go-correctness concern that templates
shouldn't be reasoning about, and the helper keeps the rule in one
well-named function — same recipe as the `…ResponseBody` rename
that obviated `SkipTypeDeclaration`.

`StrictTypeDeclaration` is a structural rewrite — appropriate when
the next change to the strict templates would otherwise add a
sixth declaration form, or when readers are clearly bouncing
between the template chain and the underlying schema state to
follow the logic.
