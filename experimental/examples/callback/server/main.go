package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"

	treefarm "github.com/oapi-codegen/oapi-codegen/experimental/examples/callback"
)

// TreeFarm implements treefarm.ServerInterface.
type TreeFarm struct {
	initiator *treefarm.CallbackInitiator
}

var _ treefarm.ServerInterface = (*TreeFarm)(nil)

func NewTreeFarm() *TreeFarm {
	initiator, err := treefarm.NewCallbackInitiator()
	if err != nil {
		log.Fatalf("Failed to create callback initiator: %v", err)
	}
	return &TreeFarm{initiator: initiator}
}

func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(treefarm.Error{
		Code:    int32(code),
		Message: message,
	})
}

func (tf *TreeFarm) PlantTree(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received PlantTree request from %s", r.RemoteAddr)

	var req treefarm.TreePlantingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.CallbackURL == "" {
		sendError(w, http.StatusBadRequest, "callbackUrl is required")
		return
	}

	id := uuid.New()

	log.Printf("Accepted tree planting: id=%s kind=%q location=%q callbackUrl=%q",
		id, req.Kind, req.Location, req.CallbackURL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(treefarm.TreeWithID{
		Location: req.Location,
		Kind:     req.Kind,
		ID:       id,
	})

	go tf.plantAndNotify(id, req)
}

func (tf *TreeFarm) plantAndNotify(id uuid.UUID, req treefarm.TreePlantingRequest) {
	delay := time.Duration(1+rand.IntN(5)) * time.Second
	log.Printf("Planting tree %s (kind=%q, location=%q) — will complete in %s",
		id, req.Kind, req.Location, delay)

	time.Sleep(delay)

	result := treefarm.TreePlantedJSONRequestBody{
		ID:       id,
		Kind:     req.Kind,
		Location: req.Location,
		Success:  true,
	}

	log.Printf("Tree %s planted, invoking callback at %s", id, req.CallbackURL)

	resp, err := tf.initiator.TreePlanted(context.Background(), req.CallbackURL, result)
	if err != nil {
		log.Printf("Callback to %s failed: %v", req.CallbackURL, err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	log.Printf("Callback to %s returned status %d", req.CallbackURL, resp.StatusCode)
}

func main() {
	port := flag.String("port", "8080", "Port for HTTP server")
	flag.Parse()

	farm := NewTreeFarm()

	mux := http.NewServeMux()
	treefarm.HandlerFromMux(farm, mux)

	// Wrap with request logging.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		mux.ServeHTTP(w, r)
	})

	addr := net.JoinHostPort("0.0.0.0", *port)
	log.Printf("Tree Farm server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
