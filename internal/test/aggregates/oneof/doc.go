// Package aggregatesoneof exercises oneOf code generation: discriminator-driven unions
// (Discriminator/ValueByDiscriminator + As/From/Merge helpers), union response marshaling
// under a strict server, and the comprehensive oneOf+anyOf marshaling cases folded from
// the components.yaml kitchen sink.
//
// Folds in:
//   - issues/issue-1530 (discriminator)            -> config_discriminator.yaml / spec_discriminator.yaml
//   - issues/issue-970  (union response, strict)   -> config_union.yaml / spec_union.yaml
//   - components/components.yaml (oneOf+anyOf)      -> config_components.yaml / spec_components.yaml
package aggregatesoneof

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_discriminator.yaml spec_discriminator.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_union.yaml spec_union.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_components.yaml spec_components.yaml
