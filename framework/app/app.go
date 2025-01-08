package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	Router *chi.Mux
}

func New() *App {
	r := chi.NewRouter()

	// Default middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	return &App{
		Router: r,
	}
}

func (a *App) Mount(pattern string, handler chi.Router) {
	a.Router.Mount(pattern, handler)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Router.ServeHTTP(w, r)
}
