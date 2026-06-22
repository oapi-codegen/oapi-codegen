// Package parametersroundtrip — HARNESS: table-driven parameter-binding roundtrip across echo/echov5/chi/gin/fiber/gorilla/iris/stdhttp + client over one shared spec. Replaces the 9 per-router dirs.
package parametersroundtrip

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
