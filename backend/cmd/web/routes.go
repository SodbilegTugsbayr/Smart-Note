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
		r.Post("/logout", logout)
		r.Post("/upload", uploadFile)

		r.Route("/ai", func(r chi.Router) {
			r.Post("/process-notes", processCourseNotes)
			r.Post("/generate-flashcards", generateCourseFlashcards)
			r.Post("/chat", askCourseChat)
		})
		r.Post("/flashcards", addFlashcard)

		r.Route("/course", func(r chi.Router) {
			r.Get("/", getUserCourses)
			r.Post("/", saveCourse)
			r.With(setChosenCourse).Route("/{CourseID}", func(r chi.Router) {
				r.Get("/", getCourse)
				r.Patch("/", updateCourse)
				r.Delete("/", deleteCourse)
				r.Post("/notes", createCourseNote)
			})
		})

		r.Route("/courses", func(r chi.Router) {
			r.With(setChosenCourse).Route("/{CourseID}", func(r chi.Router) {
				r.Get("/", getCourse)
				r.Patch("/", updateCourse)
				r.Delete("/", deleteCourse)
				r.Post("/notes", createCourseNote)
			})
		})

		r.Route("/notes", func(r chi.Router) {
			r.With(setChosenNote).Route("/{NoteID}", func(r chi.Router) {
				r.Patch("/", updateNote)
				r.Delete("/", deleteNote)
				r.Post("/file", uploadNoteFile)
				r.Get("/quizzes", getNoteQuizzes)
			})
		})

		r.With(requireAdmin).Route("/users", func(r chi.Router) {
			r.Get("/", getUsers)
			r.With(setChosenUser).Route("/{UserID}", func(r chi.Router) {
				r.Get("/", getUser)
				r.Put("/", editUser)
				r.Delete("/", deleteUser)
			})
		})
	})

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, app.Config.FilePath))
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
