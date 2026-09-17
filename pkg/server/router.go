package server

import (
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter returns the HTTP handler: /api/v1/* plus embedded UI fallback.
func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"ok":true}`)
		})
		activeAPI.mountConfig(r)
		activeAPI.mountSync(r)
		activeAPI.mountDocs(r)
		activeAPI.mountTasks(r)
	})

	ui, err := DistFS()
	if err != nil {
		r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "ui assets unavailable", http.StatusInternalServerError)
		})
		return r
	}
	fileServer := http.FileServer(http.FS(ui))
	r.Handle("/*", spaHandler(ui, fileServer))
	return r
}

func spaHandler(root fs.FS, files http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(root, p); err != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
			files.ServeHTTP(w, r)
			return
		}
		files.ServeHTTP(w, r)
	})
}
