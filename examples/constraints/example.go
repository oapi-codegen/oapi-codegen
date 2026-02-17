package constraints

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Example implementation showing how to use the generated constraint constants for both schema types and inline parameters
type Server struct{}

func NewServer() *Server {
	return &Server{}
}

// ListUsers demonstrates using constraint constants from inline parameters
func (s *Server) ListUsers(ctx echo.Context, params ListUsersParams) error {
	// Validate limit parameter using generated constants
	if params.Limit != nil {
		if *params.Limit < ListUsersLimitMinimum || *params.Limit > ListUsersLimitMaximum {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("Limit must be between %d and %d", ListUsersLimitMinimum, ListUsersLimitMaximum))
		}
	} else {
		// Use the generated default constant
		defaultLimit := ListUsersLimitDefault
		params.Limit = &defaultLimit
	}

	// Validate offset parameter
	if params.Offset != nil {
		if *params.Offset < ListUsersOffsetMinimum {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("Offset must be at least %d", ListUsersOffsetMinimum))
		}
	}

	// Validate search parameter length
	if params.Search != nil {
		searchLen := uint64(len(*params.Search))
		if searchLen < ListUsersSearchMinLength {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("Search must be at least %d characters", ListUsersSearchMinLength))
		}
		if searchLen > ListUsersSearchMaxLength {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("Search must not exceed %d characters", ListUsersSearchMaxLength))
		}
	}

	// Validate minScore parameter
	if params.MinScore != nil {
		if *params.MinScore < ListUsersMinScoreMinimum || *params.MinScore > ListUsersMinScoreMaximum {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("MinScore must be between %.1f and %.1f", ListUsersMinScoreMinimum, ListUsersMinScoreMaximum))
		}
	}

	fmt.Printf("Listing users with limit=%d, offset=%d\n", *params.Limit, *params.Offset)

	// Return mock data (implementation not shown)
	users := []User{}
	return ctx.JSON(http.StatusOK, users)
}

// GetUser demonstrates using constraint constants from path parameters
func (s *Server) GetUser(ctx echo.Context, userId string) error {
	// Validate userId length using generated constants
	userIdLen := uint64(len(userId))
	if userIdLen < GetUserUserIdMinLength || userIdLen > GetUserUserIdMaxLength {
		return echo.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("UserId must be exactly %d characters", GetUserUserIdMinLength))
	}

	fmt.Printf("Getting user: %s\n", userId)

	// Return mock data (implementation not shown)
	user := User{Username: UsernameDefault, Age: AgeDefault}
	return ctx.JSON(http.StatusOK, user)
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

	fmt.Println("=== Schema Constraint Constants ===")
	fmt.Printf("Age range: %d-%d (default: %d)\n", AgeMinimum, AgeMaximum, AgeDefault)
	fmt.Printf("Username length: %d-%d (default: %s)\n", UsernameMinLength, UsernameMaxLength, UsernameDefault)
	fmt.Printf("Port range: %d-%d (default: %d)\n", PortMinimum, PortMaximum, PortDefault)
	fmt.Printf("Score range: %.1f-%.1f (default: %.1f)\n", UserScoreMinimum, UserScoreMaximum, UserScoreDefault)
	fmt.Printf("Tags count: %d-%d\n", UserTagsMinItems, UserTagsMaxItems)
	fmt.Printf("Is active default: %v\n", IsActiveDefault)

	fmt.Println("\n=== Inline Parameter Constraint Constants ===")
	fmt.Printf("ListUsers limit: %d-%d (default: %d)\n", ListUsersLimitMinimum, ListUsersLimitMaximum, ListUsersLimitDefault)
	fmt.Printf("ListUsers offset: minimum %d\n", ListUsersOffsetMinimum)
	fmt.Printf("ListUsers search length: %d-%d\n", ListUsersSearchMinLength, ListUsersSearchMaxLength)
	fmt.Printf("ListUsers minScore: %.1f-%.1f\n", ListUsersMinScoreMinimum, ListUsersMinScoreMaximum)
	fmt.Printf("GetUser userId length: %d-%d\n", GetUserUserIdMinLength, GetUserUserIdMaxLength)

	e.Logger.Fatal(e.Start(":8080"))
}
