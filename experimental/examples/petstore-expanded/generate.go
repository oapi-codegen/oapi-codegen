//go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen -config models.config.yaml petstore-expanded.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen -config stdhttp/server/server.config.yaml -output stdhttp/server/server.gen.go petstore-expanded.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen -config chi/server/server.config.yaml -output chi/server/server.gen.go petstore-expanded.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen -config echo-v4/server/server.config.yaml -output echo-v4/server/server.gen.go petstore-expanded.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen -config echo/server/server.config.yaml -output echo/server/server.gen.go petstore-expanded.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen -config gin/server/server.config.yaml -output gin/server/server.gen.go petstore-expanded.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen -config gorilla/server/server.config.yaml -output gorilla/server/server.gen.go petstore-expanded.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen -config fiber/server/server.config.yaml -output fiber/server/server.gen.go petstore-expanded.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen-exp/experimental/cmd/oapi-codegen -config iris/server/server.config.yaml -output iris/server/server.gen.go petstore-expanded.yaml

package petstore
