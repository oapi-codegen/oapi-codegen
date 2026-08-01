// Package optionstypealiases exercises output-options.disable-type-aliases-for-type:
// when set to ["array"], array schemas that would normally emit a type alias (e.g.
// type Example = []MyItem) are instead emitted as distinct named types
// (type Example []MyItem), allowing methods to be defined on them.
package optionstypealiases

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
