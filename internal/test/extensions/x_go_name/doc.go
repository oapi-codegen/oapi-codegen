// Package extensionsxgoname verifies the x-go-name extension renames generated Go
// identifiers (RenameMe -> NewName) and that $refs resolve to the new name, including
// x-go-name on a response and a request body.
//
// components/components.yaml: RenameMe, ReferenceToRenameMe, ResponseObject, RequestBody.
package extensionsxgoname

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
