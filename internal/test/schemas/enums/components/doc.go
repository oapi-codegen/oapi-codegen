// Package schemasenumscomponents holds the enum cases folded from the components.yaml
// kitchen sink (conflict-prefixing, numeric, allOf unions, edge-case values). Compile-only.
// Isolated from the parent schemas/enums package to avoid enum-constant name collisions.
//
// components/components.yaml: Enum1-5, EnumUnion, EnumUnion2, FunnyValues.
package schemasenumscomponents

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
