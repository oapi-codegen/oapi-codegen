package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"

	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/authenticated-api/stdhttp/api"
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

func CreateMiddleware(v JWSValidator) (func(next http.Handler) http.Handler, error) {
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

	return validator, nil
}

// Ensure that we implement the server interface
var _ api.ServerInterface = (*server)(nil)

func (s *server) ListThings(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(things)
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

func (s *server) AddThing(w http.ResponseWriter, r *http.Request) {
	// This handler will only be called when the JWT is valid and the JWT contains
	// the scopes required.
	bodyBytes, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("could not bind request body"))
		return
	}

	var thing api.Thing
	err = json.Unmarshal(bodyBytes, &thing)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("could not bind request body"))
		return
	}

	s.Lock()
	defer s.Unlock()

	s.things[s.lastID] = thing
	thingWithId := api.ThingWithID{
		Name: thing.Name,
		Id:   s.lastID,
	}
	s.lastID++

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(thingWithId)
}
