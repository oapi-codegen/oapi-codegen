# issue_1378 — Path-level $ref to external file

https://github.com/oapi-codegen/oapi-codegen/issues/1378

This V2 test exercised multi-package code generation where one spec (`foo-service.yaml`) re-exported another spec's path using a path-level `$ref`:

```yaml
paths:
  /bionicle/{name}:
    $ref: "bionicle.yaml#/paths/~1bionicle~1{name}"
```

This pattern cannot be supported in oapi-codegen v3 because the underlying parser (pb33f/libopenapi) does not resolve external path-item `$ref` references. V2 used kin-openapi, which flattened all external refs (including path items) into a single in-memory document before code generation.

The OpenAPI 3.0 spec allows `$ref` at the path-item level (section 4.8.9), but it is an uncommon pattern. Users who need this can inline the referenced path operations directly.
