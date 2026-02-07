package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"

	doorbadge "github.com/oapi-codegen/oapi-codegen-exp/experimental/examples/webhook"
)

// WebhookReceiver implements doorbadge.WebhookReceiverInterface.
type WebhookReceiver struct{}

var _ doorbadge.WebhookReceiverInterface = (*WebhookReceiver)(nil)

func (wr *WebhookReceiver) HandleEnterEventWebhook(w http.ResponseWriter, r *http.Request) {
	var person doorbadge.Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		log.Printf("Error decoding enter event: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("ENTER: %s", person.Name)
	w.WriteHeader(http.StatusOK)
}

func (wr *WebhookReceiver) HandleExitEventWebhook(w http.ResponseWriter, r *http.Request) {
	var person doorbadge.Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		log.Printf("Error decoding exit event: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("EXIT:  %s", person.Name)
	w.WriteHeader(http.StatusOK)
}

func registerWebhook(client *http.Client, serverAddr, kind, url string) (uuid.UUID, error) {
	body, err := json.Marshal(doorbadge.WebhookRegistration{URL: url})
	if err != nil {
		return uuid.UUID{}, err
	}

	resp, err := client.Post(
		serverAddr+"/api/webhook/"+kind,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return uuid.UUID{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var regResp doorbadge.WebhookRegistrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return uuid.UUID{}, err
	}
	return regResp.ID, nil
}

func deregisterWebhook(client *http.Client, serverAddr string, id uuid.UUID) error {
	req, err := http.NewRequest(http.MethodDelete, serverAddr+"/api/webhook/"+id.String(), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

func main() {
	serverAddr := flag.String("server", "http://localhost:8080", "Badge reader server address")
	duration := flag.Duration("duration", 30*time.Second, "How long to listen for events")
	flag.Parse()

	// Start the webhook receiver on an ephemeral port.
	receiver := &WebhookReceiver{}

	mux := http.NewServeMux()
	mux.Handle("POST /enter", doorbadge.EnterEventWebhookHandler(receiver, nil))
	mux.Handle("POST /exit", doorbadge.ExitEventWebhookHandler(receiver, nil))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	callbackPort := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://localhost:%d", callbackPort)
	log.Printf("Webhook receiver listening on port %d", callbackPort)

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Webhook server stopped: %v", err)
		}
	}()

	// Register webhooks for both event kinds.
	client := &http.Client{}

	kinds := [2]string{"enterEvent", "exitEvent"}
	urls := [2]string{baseURL + "/enter", baseURL + "/exit"}
	var registrationIDs [2]uuid.UUID

	for i, kind := range kinds {
		id, err := registerWebhook(client, *serverAddr, kind, urls[i])
		if err != nil {
			log.Fatalf("Failed to register %s webhook: %v", kind, err)
		}
		registrationIDs[i] = id
		log.Printf("Registered %s webhook: id=%s url=%s", kind, id, urls[i])
	}

	log.Printf("Listening for events for %s...", *duration)

	// Wait for the specified duration.
	time.Sleep(*duration)

	// Deregister webhooks cleanly.
	log.Printf("Duration elapsed, deregistering webhooks...")
	for i, id := range registrationIDs {
		if err := deregisterWebhook(client, *serverAddr, id); err != nil {
			log.Printf("Failed to deregister %s webhook: %v", kinds[i], err)
			continue
		}
		log.Printf("Deregistered %s webhook: id=%s", kinds[i], id)
	}

	// Shut down the local webhook server.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	log.Printf("Done!")
}
