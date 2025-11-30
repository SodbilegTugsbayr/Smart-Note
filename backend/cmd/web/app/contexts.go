package app

type contextKey string

const (
	ContextKeyIsAuthenticated = contextKey("isAuthenticated")
	ContextKeyAuthUser        = contextKey("authenticatedUser")
	ContextKeyChosenUser      = contextKey("chosenUser")
	ContextKeyChosenCategory  = contextKey("chosenCategory")
	ContextKeyChosenAdvert    = contextKey("chosenAdvert")
	ContextKeyChosenBanner    = contextKey("chosenBanner")
)
