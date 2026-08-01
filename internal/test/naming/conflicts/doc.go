// Package namingconflicts exercises name collision resolution across all OpenAPI
// component sections (schemas, parameters, requestBodies, responses, headers),
// client wrapper name conflicts, x-go-name / x-go-type / x-go-type-name
// extension interactions with the resolver, and avoidance of package-name
// collisions in field comment text (import-name grab).
//
// Folds in:
//   - name_conflict_resolution (comprehensive cross-section collision patterns A–M)
//   - issues/issue-grab_import_names (pkg names in comments must not become imports;
//     tested via in-memory codegen.Generate call, so no separate gen triple needed)
package namingconflicts

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
