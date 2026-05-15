package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router *chi.Mux
}

func New(booking *BookingHandler, tracking *TrackingHandler, handling *HandlingHandler) *Server {
	s := &Server{
		router: chi.NewRouter(),
	}

	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(jsonContentType)

	s.router.Route("/api", func(r chi.Router) {
		r.Post("/cargos", booking.Book)
		r.Get("/cargos", booking.List)
		r.Get("/cargos/{id}", booking.Show)
		r.Get("/cargos/{id}/routes", booking.RequestRoutes)
		r.Put("/cargos/{id}/route", booking.AssignRoute)
		r.Put("/cargos/{id}/destination", booking.ChangeDestination)

		r.Get("/locations", booking.ListLocations)

		r.Get("/track/{id}", tracking.Track)

		r.Post("/handling", handling.RegisterEvent)
	})

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, errorResponse{Error: msg})
}
