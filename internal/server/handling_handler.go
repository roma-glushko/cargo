package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/handling"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

type HandlingHandler struct {
	service *handling.Service
}

func NewHandlingHandler(service *handling.Service) *HandlingHandler {
	return &HandlingHandler{service: service}
}

func (h *HandlingHandler) RegisterEvent(w http.ResponseWriter, r *http.Request) {
	var req registerEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	completionTime, err := time.Parse(timeFormat, req.CompletionTime)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid completionTime format, use RFC3339")
		return
	}

	eventType, ok := cargo.ParseEventType(req.Type)
	if !ok {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("unknown event type: %s", req.Type))
		return
	}

	var errors []string
	for _, tid := range req.TrackingIDs {
		err := h.service.RegisterHandlingEvent(
			r.Context(),
			completionTime,
			cargo.TrackingID(tid),
			voyage.Number(req.VoyageNumber),
			location.UNLocode(req.Location),
			eventType,
		)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", tid, err.Error()))
		}
	}

	if len(errors) > 0 {
		respondJSON(w, http.StatusBadRequest, map[string]any{"errors": errors})
		return
	}

	w.WriteHeader(http.StatusCreated)
}
