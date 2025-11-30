package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/peakdot/go-nuxt-example/backend/cmd/web/app"
	"github.com/peakdot/go-nuxt-example/backend/pkg/common/oapi"
	"github.com/peakdot/go-nuxt-example/backend/pkg/userman"
)

type APIError struct {
	Code    int
	Message string
	Status  int
}

var (
	ErrAccountNotFound = APIError{400, "Account not Found", http.StatusBadRequest}
	ErrInvalidEmail    = APIError{401, "Invaid email", http.StatusBadRequest}
)

func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedUserID := app.Session.GetInt(r, "auth_user_id")
		if loggedUserID <= 0 {
			ctx := context.WithValue(r.Context(), app.ContextKeyIsAuthenticated, false)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		user, err := app.Users.GetByID(loggedUserID)
		if err != nil {
			if errors.Is(err, userman.ErrNotFound) {
				ctx := context.WithValue(r.Context(), app.ContextKeyIsAuthenticated, false)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				oapi.ServerError(w, err)
			}
			return
		}

		ctx := context.WithValue(r.Context(), app.ContextKeyIsAuthenticated, true)
		ctx = context.WithValue(ctx, app.ContextKeyAuthUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth, _ := r.Context().Value(app.ContextKeyIsAuthenticated).(bool)
		if !isAuth {
			oapi.ClientError(w, http.StatusUnauthorized)
			return
		}

		user := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)
		if !user.IsVerified {
			oapi.ClientError(w, http.StatusPreconditionFailed)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

		if user.Role != userman.ROLE_ADMIN {
			oapi.ClientError(w, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setChosenUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "UserID"))

		user, err := app.Users.GetByID(id)
		if err != nil {
			if errors.Is(err, userman.ErrNotFound) {
				oapi.NotFound(w)
			} else {
				oapi.ServerError(w, err)
			}
			return
		}

		ctx := context.WithValue(r.Context(), app.ContextKeyChosenUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
