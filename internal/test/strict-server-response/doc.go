package matrix

//go:generate go run generator/generate.go
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=chi/pkg1/config.yaml pkg1.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=chi/pkg2/config.yaml pkg2.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=echo/pkg1/config.yaml pkg1.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=echo/pkg2/config.yaml pkg2.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=fiber/pkg1/config.yaml pkg1.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=fiber/pkg2/config.yaml pkg2.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=gin/pkg1/config.yaml pkg1.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=gin/pkg2/config.yaml pkg2.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=gorilla/pkg1/config.yaml pkg1.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=gorilla/pkg2/config.yaml pkg2.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=iris/pkg1/config.yaml pkg1.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=iris/pkg2/config.yaml pkg2.yaml
