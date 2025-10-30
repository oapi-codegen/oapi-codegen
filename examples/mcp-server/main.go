package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	transport = flag.String("transport", "stdio", "Transport type: stdio or http")
	httpAddr  = flag.String("http-addr", ":8080", "HTTP address to listen on (only for http transport)")
)

// PetStore implements the StrictServerInterface for our pet store
type PetStore struct {
	mu   sync.RWMutex
	pets map[string]*Pet
}

func NewPetStore() *PetStore {
	return &PetStore{
		pets: make(map[string]*Pet),
	}
}

func (s *PetStore) ListPets(ctx context.Context, request ListPetsRequestObject) (ListPetsResponseObject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := 20
	if request.Params.Limit != nil {
		limit = int(*request.Params.Limit)
	}

	var pets []Pet
	for _, pet := range s.pets {
		// Filter by tag if provided
		if request.Params.Tag != nil && pet.Tag != nil {
			if *pet.Tag != *request.Params.Tag {
				continue
			}
		}
		pets = append(pets, *pet)
		if len(pets) >= limit {
			break
		}
	}

	return ListPets200JSONResponse{
		Pets: pets,
	}, nil
}

func (s *PetStore) CreatePet(ctx context.Context, request CreatePetRequestObject) (CreatePetResponseObject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a simple ID
	id := fmt.Sprintf("pet-%d", len(s.pets)+1)

	pet := &Pet{
		Id:    id,
		Name:  request.Body.Name,
		Tag:   request.Body.Tag,
		Age:   request.Body.Age,
		Breed: request.Body.Breed,
	}

	s.pets[id] = pet
	return CreatePet201JSONResponse(*pet), nil
}

func (s *PetStore) GetPet(ctx context.Context, request GetPetRequestObject) (GetPetResponseObject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pet, ok := s.pets[request.PetId]
	if !ok {
		return GetPet404Response{}, nil
	}

	return GetPet200JSONResponse(*pet), nil
}

func (s *PetStore) UpdatePet(ctx context.Context, request UpdatePetRequestObject) (UpdatePetResponseObject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pet, ok := s.pets[request.PetId]
	if !ok {
		return UpdatePet404Response{}, nil
	}

	// Update fields if provided
	if request.Body != nil {
		if request.Body.Name != nil {
			pet.Name = *request.Body.Name
		}
		if request.Body.Tag != nil {
			pet.Tag = request.Body.Tag
		}
		if request.Body.Age != nil {
			pet.Age = request.Body.Age
		}
		if request.Body.Breed != nil {
			pet.Breed = request.Body.Breed
		}
	}

	return UpdatePet200JSONResponse(*pet), nil
}

func (s *PetStore) DeletePet(ctx context.Context, request DeletePetRequestObject) (DeletePetResponseObject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.pets[request.PetId]; !ok {
		return DeletePet404Response{}, nil
	}

	delete(s.pets, request.PetId)
	return DeletePet204Response{}, nil
}

// No adapter needed! The generated RegisterMCPServer function works directly with *mcp.Server

func main() {
	flag.Parse()

	// Create our pet store implementation
	store := NewPetStore()

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "pet-store-mcp",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		Instructions: "A pet store API exposed as MCP tools. You can list, create, get, update, and delete pets.",
	})

	// Create strict MCP handler that adapts StrictServerInterface to MCPHandlerInterface
	strictHandler := NewStrictMCPHandler(store, nil)

	// Register our handlers using the generic MCPHandlerInterface
	if err := RegisterMCPTools(mcpServer, strictHandler); err != nil {
		log.Fatalf("Failed to register MCP tools: %v", err)
	}

	// Run with the appropriate transport
	ctx := context.Background()
	switch *transport {
	case "stdio":
		log.Println("Starting MCP server with stdio transport")
		if err := mcpServer.Run(ctx, &mcp.StdioTransport{}); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	case "http":
		log.Printf("Starting MCP server with HTTP transport on %s", *httpAddr)
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return mcpServer
		}, nil)
		if err := http.ListenAndServe(*httpAddr, handler); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	default:
		log.Fatalf("Unknown transport: %s (use 'stdio' or 'http')", *transport)
	}
}
