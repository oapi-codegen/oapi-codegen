package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer() *Server {
	s := &Server{}
	strictHandler := NewStrictHandler(s, nil)
	handler := Handler(strictHandler)
	s.httpServer = &http.Server{
		Handler: handler,
		Addr:    ":8080",
	}
	return s
}
func (s Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		slog.Warn("context cancelled, shutting down")
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := s.httpServer.Shutdown(ctxTimeout)
		cancel()
		if err != nil {
			slog.Error("httpServer.Shutdown() error", "error", err)
		}
	}()

	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("httpServer.ListenAndServe() error: %w", err)
	}
	return nil
}

type SObject struct {
	Time     time.Time `json:"time"`
	Sequence int       `json:"sequence"`
}

// GetStream handles GET / and will stream a JSON object every second.
func (Server) GetStream(ctx context.Context, _ GetStreamRequestObject) (GetStreamResponseObject, error) {
	r, w := io.Pipe() // creates a pipe so that we can write to the response body asynchronously
	go func() {
		defer w.Close()
		seq := 1
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				slog.Info("request context done, closing stream")
				return
			case <-ticker.C:
				content := getContent(seq)
				if _, err := w.Write(content); err != nil {
					return
				}
				if _, err := w.Write([]byte("\n")); err != nil {
					return
				}
				seq++
			}
		}
	}()
	return GetStream200TexteventStreamResponse{
		Body:          r,
		ContentLength: 0,
	}, nil
}

func getContent(seq int) []byte {
	streamObject := SObject{
		Time:     time.Now(),
		Sequence: seq,
	}
	bytes, _ := json.Marshal(streamObject)
	return bytes
}
