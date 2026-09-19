package support

// GenericResponse is a concrete generic response body used by code-generation tests.
type GenericResponse[T any] struct {
	Value T `json:"value"`
}
