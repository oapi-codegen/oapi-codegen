package server

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=config.yaml ../test-schema.yaml

// This is commented out because the server_mog.gen.go keeps changing for no good reason, and
// so, precommit checks fail. We need to regenerate this file occasionally manually.
// TODO(mromaszewicz) - figure out why this file drifts and fix it.
// go:generate go run github.com/matryer/moq -out server_moq.gen.go . ServerInterface
