package auth

import "context"

// Auth functions
type Auth interface {
	APIKey(token string) (interface{}, error)
	Basic(user, password string) (interface{}, error)
	OAuth2(token string, scopes []string) (interface{}, error)
	// AuthStore is a function that stores authentication in the context object
	Store(context.Context, interface{}) context.Context
}
