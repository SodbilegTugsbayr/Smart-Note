package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(app.Session.Enable)

	r.Get("/ping", ping)

	// Public routes
	r.With(authenticate).Route("/pub", func(r chi.Router) {
		r.Get("/logout", clearSession)

		r.Post("/echo", echo)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", login)
			r.Post("/signup", signup)
			r.Route("/google", func(r chi.Router) {
				r.Get("/login", oauthLogin(app.GoogleOAuth2))
				r.Get("/callback", oauthCallback(app.GoogleOAuth2))
			})
			r.Route("/facebook", func(r chi.Router) {
				r.Get("/login", oauthLogin(app.FacebookOAuth2))
				r.Get("/callback", oauthCallback(app.FacebookOAuth2))
			})
		})

	})

	// Authenticated routes
	r.With(authenticate, requireAuth).Route("/api", func(r chi.Router) {
		r.Get("/me", me)
		r.Post("/me", updateUserInfo)
		r.HandleFunc("/ws", app.CustomerWSConnections.Handler)
		r.Get("/logout", logout)

		// TODO: Future plan: Admin
		r.With(requireAdmin).Route("/users", func(r chi.Router) {
			r.Get("/", getUsers)
			r.Post("/", editUser)
			r.With(setChosenUser).Route("/{UserID}", func(r chi.Router) {
				r.Get("/", getUser)
				r.Put("/", editUser)
				r.Delete("/", deleteUser)
			})
		})
	})

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, app.Config.ImagePath))
	FileServer(r, "/pub/images", filesDir)

	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
