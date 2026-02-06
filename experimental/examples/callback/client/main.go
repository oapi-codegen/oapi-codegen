package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	treefarm "github.com/oapi-codegen/oapi-codegen/experimental/examples/callback"
)

// trees and cities for our planting requests
var treeKinds = []string{
	"oak", "maple", "pine", "birch", "willow",
	"cedar", "elm", "ash", "cherry", "walnut",
}

var cities = []string{
	"Providence", "Austin", "Denver", "Seattle", "Chicago",
	"Boston", "Miami", "Nashville", "Savannah", "Mountain View",
}

// CallbackReceiver implements treefarm.CallbackReceiverInterface.
type CallbackReceiver struct {
	received atomic.Int32
	total    int
	done     chan struct{}
	once     sync.Once

	mu       sync.Mutex
	ordinals map[string]int // UUID string -> 1-based planting order
}

var _ treefarm.CallbackReceiverInterface = (*CallbackReceiver)(nil)

func NewCallbackReceiver(total int) *CallbackReceiver {
	return &CallbackReceiver{
		total:    total,
		done:     make(chan struct{}),
		ordinals: make(map[string]int),
	}
}

func (cr *CallbackReceiver) Register(id string, ordinal int) {
	cr.mu.Lock()
	cr.ordinals[id] = ordinal
	cr.mu.Unlock()
}

func (cr *CallbackReceiver) HandleTreePlantedCallback(w http.ResponseWriter, r *http.Request) {
	var result treefarm.TreePlantingResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		log.Printf("Error decoding callback: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cr.mu.Lock()
	ordinal := cr.ordinals[result.ID.String()]
	cr.mu.Unlock()

	n := cr.received.Add(1)
	log.Printf("Callback %d/%d received: tree #%d success=%v", n, cr.total, ordinal, result.Success)

	w.WriteHeader(http.StatusOK)

	if int(n) >= cr.total {
		cr.once.Do(func() { close(cr.done) })
	}
}

func main() {
	serverAddr := flag.String("server", "http://localhost:8080", "Tree farm server address")
	flag.Parse()

	const numTrees = 10

	// Start callback receiver on an ephemeral port.
	receiver := NewCallbackReceiver(numTrees)

	mux := http.NewServeMux()
	mux.Handle("/tree_callback", treefarm.TreePlantedCallbackHandler(receiver, nil))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	callbackPort := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://localhost:%d/tree_callback", callbackPort)
	log.Printf("Callback receiver listening on port %d", callbackPort)

	go func() {
		if err := http.Serve(listener, mux); err != nil {
			log.Printf("Callback server stopped: %v", err)
		}
	}()

	// Send 10 tree planting requests.
	client := &http.Client{}
	for i := range numTrees {
		req := treefarm.TreePlantingRequest{
			Kind:        treeKinds[i],
			Location:    cities[i],
			CallbackURL: callbackURL,
		}

		body, err := json.Marshal(req)
		if err != nil {
			log.Fatalf("Failed to marshal request: %v", err)
		}

		resp, err := client.Post(
			*serverAddr+"/api/plant_tree",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			log.Fatalf("Failed to plant tree %d: %v", i+1, err)
		}

		var accepted treefarm.TreeWithID
		if err := json.NewDecoder(resp.Body).Decode(&accepted); err != nil {
			resp.Body.Close()
			log.Fatalf("Failed to decode response: %v", err)
		}
		resp.Body.Close()

		receiver.Register(accepted.ID.String(), i+1)
		log.Printf("Planted tree %d/%d: id=%s kind=%q location=%q",
			i+1, numTrees, accepted.ID, accepted.Kind, accepted.Location)
	}

	log.Printf("All %d trees planted, waiting for callbacks...", numTrees)

	// Wait for all callbacks.
	<-receiver.done

	log.Printf("All %d callbacks received, done!", numTrees)
}
