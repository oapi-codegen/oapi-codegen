package constraints

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Example implementation showing how to use the generated constraint constants
type Server struct{}

func NewServer() *Server {
	return &Server{}
}

// CreateUser demonstrates using the generated constraint constants for validation
func (s *Server) CreateUser(ctx echo.Context) error {
	var user User
	if err := ctx.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Use the generated constraint constants for validation!

	// Validate age constraints
	if user.Age < AgeMinimum || user.Age > AgeMaximum {
		return echo.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("Age must be between %d and %d", AgeMinimum, AgeMaximum))
	}

	// Validate username length constraints
	if uint64(len(user.Username)) < UsernameMinLength {
		return echo.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("Username must be at least %d characters", UsernameMinLength))
	}
	if uint64(len(user.Username)) > UsernameMaxLength {
		return echo.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("Username must not exceed %d characters", UsernameMaxLength))
	}

	// Validate port constraints (if provided)
	if user.Port != nil {
		if *user.Port < PortMinimum || *user.Port > PortMaximum {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("Port must be between %d and %d", PortMinimum, PortMaximum))
		}
	}

	// Validate user score constraints (if provided)
	if user.Score != nil {
		if *user.Score < UserScoreMinimum || *user.Score > UserScoreMaximum {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("Score must be between %.1f and %.1f", UserScoreMinimum, UserScoreMaximum))
		}
	}

	// Validate tags constraints (if provided)
	if user.Tags != nil {
		if uint64(len(*user.Tags)) < UserTagsMinItems {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("Must provide at least %d tags", UserTagsMinItems))
		}
		if uint64(len(*user.Tags)) > UserTagsMaxItems {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("Cannot exceed %d tags", UserTagsMaxItems))
		}
	}

	// Use default values where appropriate
	if user.IsActive == nil {
		defaultActive := IsActiveDefault
		user.IsActive = &defaultActive
	}
	if user.Port == nil {
		defaultPort := PortDefault
		user.Port = &defaultPort
	}
	if user.Score == nil {
		defaultScore := UserScoreDefault
		user.Score = &defaultScore
	}

	// Create the user (implementation not shown)
	fmt.Printf("Creating user: %+v\n", user)

	return ctx.JSON(http.StatusCreated, user)
}

// Example usage in main function
func Example() {
	e := echo.New()
	server := NewServer()
	RegisterHandlers(e, server)

	// The constraints are available as typed constants:
	fmt.Printf("Age range: %d-%d (default: %d)\n", AgeMinimum, AgeMaximum, AgeDefault)
	fmt.Printf("Username length: %d-%d (default: %s)\n", UsernameMinLength, UsernameMaxLength, UsernameDefault)
	fmt.Printf("Port range: %d-%d (default: %d)\n", PortMinimum, PortMaximum, PortDefault)
	fmt.Printf("Score range: %.1f-%.1f (default: %.1f)\n", UserScoreMinimum, UserScoreMaximum, UserScoreDefault)
	fmt.Printf("Tags count: %d-%d\n", UserTagsMinItems, UserTagsMaxItems)
	fmt.Printf("Is active default: %v\n", IsActiveDefault)

	e.Logger.Fatal(e.Start(":8080"))
}
