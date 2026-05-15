package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/roma-glushko/cargo/internal/booking"
	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
)

type BookingHandler struct {
	service   *booking.Service
	cargos    cargo.Repository
	locations location.Repository
}

func NewBookingHandler(service *booking.Service, cargos cargo.Repository, locations location.Repository) *BookingHandler {
	return &BookingHandler{
		service:   service,
		cargos:    cargos,
		locations: locations,
	}
}

func (h *BookingHandler) Book(w http.ResponseWriter, r *http.Request) {
	var req bookCargoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	deadline, err := time.Parse(timeFormat, req.Deadline)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid deadline format, use RFC3339")
		return
	}

	id, err := h.service.BookNewCargo(r.Context(), location.UNLocode(req.Origin), location.UNLocode(req.Destination), deadline)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, bookCargoResponse{TrackingID: string(id)})
}

func (h *BookingHandler) List(w http.ResponseWriter, r *http.Request) {
	cargos, err := h.cargos.FindAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]cargoResponse, len(cargos))
	for i, c := range cargos {
		result[i] = cargoToResponse(c)
	}

	respondJSON(w, http.StatusOK, result)
}

func (h *BookingHandler) Show(w http.ResponseWriter, r *http.Request) {
	id := cargo.TrackingID(chi.URLParam(r, "id"))

	c, err := h.cargos.Find(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "cargo not found")
		return
	}

	respondJSON(w, http.StatusOK, cargoToResponse(c))
}

func (h *BookingHandler) RequestRoutes(w http.ResponseWriter, r *http.Request) {
	id := cargo.TrackingID(chi.URLParam(r, "id"))

	itineraries, err := h.service.RequestPossibleRoutes(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	var result []itineraryResponse
	for _, itin := range itineraries {
		var legs []legDTO
		for _, leg := range itin.Legs {
			legs = append(legs, legToDTO(leg))
		}
		result = append(result, itineraryResponse{Legs: legs})
	}

	respondJSON(w, http.StatusOK, result)
}

func (h *BookingHandler) AssignRoute(w http.ResponseWriter, r *http.Request) {
	id := cargo.TrackingID(chi.URLParam(r, "id"))

	var req assignRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	itin := dtoToItinerary(req)
	if err := h.service.AssignToRoute(r.Context(), id, itin); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BookingHandler) ChangeDestination(w http.ResponseWriter, r *http.Request) {
	id := cargo.TrackingID(chi.URLParam(r, "id"))

	var req changeDestinationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.ChangeDestination(r.Context(), id, location.UNLocode(req.Destination)); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BookingHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := h.locations.FindAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]locationResponse, len(locations))
	for i, loc := range locations {
		result[i] = locationToResponse(loc)
	}

	respondJSON(w, http.StatusOK, result)
}
