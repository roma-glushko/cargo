package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/roma-glushko/cargo/internal/cargo"
)

type TrackingHandler struct {
	cargos cargo.Repository
	events cargo.HandlingEventRepository
}

func NewTrackingHandler(cargos cargo.Repository, events cargo.HandlingEventRepository) *TrackingHandler {
	return &TrackingHandler{
		cargos: cargos,
		events: events,
	}
}

func (h *TrackingHandler) Track(w http.ResponseWriter, r *http.Request) {
	id := cargo.TrackingID(chi.URLParam(r, "id"))

	c, err := h.cargos.Find(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "cargo not found")
		return
	}

	history, err := h.events.LookupHandlingHistory(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := trackingResponse{
		TrackingID:    string(c.TrackingID),
		StatusText:    formatStatusText(c),
		Destination:   string(c.RouteSpecification.Destination),
		IsMisdirected: c.Delivery.Misdirected,
	}

	if !c.Delivery.ETA.IsZero() {
		resp.ETA = c.Delivery.ETA.Format(timeFormat)
	}

	if act := c.Delivery.NextExpectedActivity; act != nil {
		resp.NextExpectedActivity = formatActivity(act)
	}

	for _, e := range history.Events {
		resp.Events = append(resp.Events, handlingEventResponse{
			Location:       string(e.Location),
			CompletionTime: e.CompletionTime.Format(timeFormat),
			Type:           e.Type.String(),
			VoyageNumber:   string(e.VoyageNumber),
			IsExpected:     c.Itinerary.IsExpected(e),
			Description:    formatEventDescription(e),
		})
	}

	respondJSON(w, http.StatusOK, resp)
}

func formatStatusText(c *cargo.Cargo) string {
	switch c.Delivery.TransportStatus {
	case cargo.NotReceived:
		return "Not received"
	case cargo.InPort:
		return fmt.Sprintf("In port at %s", c.Delivery.LastKnownLocation)
	case cargo.OnboardCarrier:
		return fmt.Sprintf("Onboard voyage %s", c.Delivery.CurrentVoyage)
	case cargo.Claimed:
		return "Claimed"
	default:
		return "Unknown"
	}
}

func formatActivity(act *cargo.HandlingActivity) string {
	switch act.Type {
	case cargo.Receive:
		return fmt.Sprintf("Receive cargo at %s", act.Location)
	case cargo.Load:
		return fmt.Sprintf("Load cargo onto voyage %s at %s", act.VoyageNumber, act.Location)
	case cargo.Unload:
		return fmt.Sprintf("Unload cargo from voyage %s at %s", act.VoyageNumber, act.Location)
	case cargo.Claim:
		return fmt.Sprintf("Claim cargo at %s", act.Location)
	case cargo.Customs:
		return fmt.Sprintf("Customs at %s", act.Location)
	default:
		return ""
	}
}

func formatEventDescription(e cargo.HandlingEvent) string {
	switch e.Type {
	case cargo.Receive:
		return fmt.Sprintf("Received at %s", e.Location)
	case cargo.Load:
		return fmt.Sprintf("Loaded onto voyage %s at %s", e.VoyageNumber, e.Location)
	case cargo.Unload:
		return fmt.Sprintf("Unloaded from voyage %s at %s", e.VoyageNumber, e.Location)
	case cargo.Claim:
		return fmt.Sprintf("Claimed at %s", e.Location)
	case cargo.Customs:
		return fmt.Sprintf("Customs at %s", e.Location)
	default:
		return ""
	}
}
