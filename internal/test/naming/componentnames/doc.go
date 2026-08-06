// Package namingcomponentnames exercises output-options.component-names: the
// renaming of the fixed, spec-independent package-level identifiers that
// oapi-codegen emits.
//
// Two configurations generate from one spec into this one Go package, which
// is only possible because component-names.prefix renames the unexported
// names (swaggerSpec, rawSpec, decodeSpec, decodeSpecCached, strictHandler)
// as well as the exported ones. config.yaml emits models, a client, a
// std-http server, a strict server and the embedded spec under the PetStore
// prefix with several root overrides; config_chi.yaml emits a chi server and
// a second embedded spec under the Admin prefix.
//
// Between them the two configurations declare every fixed name the templates
// can emit, so the committed .gen.go files are the stale-literal detector for
// this feature: a name a template still hardcodes either fails to compile
// (undefined identifier) or shows up as a redeclaration across the two runs.
package namingcomponentnames

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_chi.yaml spec.yaml
