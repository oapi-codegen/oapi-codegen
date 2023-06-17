// This is an example of how to reference models of one api specification from another.
// See https://github.com/deepmap/oapi-codegen/issues/1093
package import_mapping

//go:generate oapi-codegen --config parent.cfg.yaml parent.api.yaml
//go:generate oapi-codegen --config child.cfg.yaml child.api.yaml
