package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Stratoscale/swagger/example/restapi"
	"github.com/go-openapi/runtime/middleware"
)

// User described a user that uses the API
type User struct {
	ID   int
	Role string
}

// Token parses and validates a token and return the logged in user
func Token(token string) (interface{}, error) {
	if token == "" {
		return nil, nil // unauthorized
	}

	// In a real authentication, here we should actually validate that the token is valid
	var user User
	err := json.Unmarshal([]byte(token), &user)
	return &user, err
}

// Request enforce policy on a given request
func Request(req *http.Request) error {
	var (
		route = middleware.MatchedRouteFrom(req)
		user  = FromContext(req.Context())
	)

	for _, auth := range route.Authenticators {
		scopes := auth.Scopes["token"]

		if len(scopes) == 0 {
			return nil // The token is valid for any user role
		}

		// Check if any of the scopes is the same as the user's role
		for _, scope := range scopes {
			if scope == user.Role {
				return nil
			}
		}
	}
	return fmt.Errorf("forbidden")
}

// FromContext extract the user from the context
func FromContext(ctx context.Context) *User {
	v := ctx.Value(restapi.AuthKey)
	if v == nil {
		return nil
	}
	return v.(*User)
}
