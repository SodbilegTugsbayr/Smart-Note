package app

type contextKey string

const (
	ContextKeyIsAuthenticated = contextKey("isAuthenticated")
	ContextKeyAuthUser        = contextKey("authenticatedUser")
	ContextKeyChosenUser      = contextKey("chosenUser")
	ContextKeyChosenCourse    = contextKey("chosenCourse")
	ContextKeyChosenNote      = contextKey("chosenNote")
	ContextKeyChosenCategory  = contextKey("chosenCategory")
	ContextKeyChosenAdvert    = contextKey("chosenAdvert")
	ContextKeyChosenBanner    = contextKey("chosenBanner")
)
