// This is an example of how to reference models of one api specification from another.
// See https://github.com/oapi-codegen/oapi-codegen/issues/1093
package issue1093

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config parent.cfg.yaml parent.api.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config child.cfg.yaml child.api.yaml
