package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
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
		fmt.Println("context cancelled, shutting down")
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

// (GET /ping)
func (Server) GetStream(ctx context.Context, _ GetStreamRequestObject) (GetStreamResponseObject, error) {
	// 10% chance of returning an error
	if rand.IntN(10) == 0 {
		return nil, errors.New("random error")
	}
	fmt.Println("GetStream invoked")
	defer fmt.Println("GetStream returneds")
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		defer fmt.Println("goro done")
		seq := 1
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("goro finished up")
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
				fmt.Println("output sent")
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
