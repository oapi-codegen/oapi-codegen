package server

import (
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/deepmap/oapi-codegen/examples/authenticated-api/echo/api"
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
)

type server struct {
	sync.RWMutex
	lastID int64
	things map[int64]api.Thing
}

func NewServer() *server {
	return &server{
		lastID: 0,
		things: make(map[int64]api.Thing),
	}
}

func CreateMiddleware(v JWSValidator) ([]echo.MiddlewareFunc, error) {
	spec, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("loading spec: %w", err)
	}

	validator := middleware.OapiRequestValidatorWithOptions(spec,
		&middleware.Options{
			Options: openapi3filter.Options{
				AuthenticationFunc: NewAuthenticator(v),
			},
		})

	return []echo.MiddlewareFunc{validator}, nil
}

// Ensure that we implement the server interface
var _ api.ServerInterface = (*server)(nil)

func (s *server) ListThings(ctx echo.Context) error {
	// This handler will only be called when a valid JWT is presented for
	// access.
	s.RLock()

	thingKeys := make([]int64, 0, len(s.things))
	for key := range s.things {
		thingKeys = append(thingKeys, key)
	}
	sort.Sort(int64s(thingKeys))

	things := make([]api.ThingWithID, 0, len(s.things))

	for _, key := range thingKeys {
		thing := s.things[key]
		things = append(things, api.ThingWithID{
			Id:   key,
			Name: thing.Name,
		})
	}

	s.RUnlock()

	return ctx.JSON(http.StatusOK, things)
}

type int64s []int64

func (in int64s) Len() int {
	return len(in)
}

func (in int64s) Less(i, j int) bool {
	return in[i] < in[j]
}

func (in int64s) Swap(i, j int) {
	in[i], in[j] = in[j], in[i]
}

var _ sort.Interface = (int64s)(nil)

func (s *server) AddThing(ctx echo.Context) error {
	// This handler will only be called when the JWT is valid and the JWT contains
	// the scopes required.
	var thing api.Thing
	err := ctx.Bind(&thing)
	if err != nil {
		return returnError(ctx, http.StatusBadRequest, "could not bind request body")
	}

	s.Lock()
	defer s.Unlock()

	s.things[s.lastID] = thing
	thingWithId := api.ThingWithID{
		Name: thing.Name,
		Id:   s.lastID,
	}
	s.lastID++

	return ctx.JSON(http.StatusCreated, thingWithId)
}

func returnError(ctx echo.Context, code int, message string) error {
	errResponse := api.Error{
		Code:    int32(code),
		Message: message,
	}
	return ctx.JSON(code, errResponse)
}
