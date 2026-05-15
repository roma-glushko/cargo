package server

import (
	"time"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

// --- Requests ---

type bookCargoRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Deadline    string `json:"arrivalDeadline"`
}

type assignRouteRequest struct {
	Legs []legDTO `json:"legs"`
}

type changeDestinationRequest struct {
	Destination string `json:"destination"`
}

type registerEventRequest struct {
	CompletionTime string   `json:"completionTime"`
	TrackingIDs    []string `json:"trackingIds"`
	Type           string   `json:"type"`
	Location       string   `json:"unLocode"`
	VoyageNumber   string   `json:"voyageNumber,omitempty"`
}

// --- Responses ---

type bookCargoResponse struct {
	TrackingID string `json:"trackingId"`
}

type cargoResponse struct {
	TrackingID      string   `json:"trackingId"`
	Origin          string   `json:"origin"`
	Destination     string   `json:"destination"`
	ArrivalDeadline string   `json:"arrivalDeadline"`
	Misrouted       bool     `json:"misrouted"`
	Routed          bool     `json:"routed"`
	Legs            []legDTO `json:"legs,omitempty"`
}

type legDTO struct {
	VoyageNumber string `json:"voyageNumber"`
	From         string `json:"from"`
	To           string `json:"to"`
	LoadTime     string `json:"loadTime"`
	UnloadTime   string `json:"unloadTime"`
}

type itineraryResponse struct {
	Legs []legDTO `json:"legs"`
}

type trackingResponse struct {
	TrackingID           string                  `json:"trackingId"`
	StatusText           string                  `json:"statusText"`
	Destination          string                  `json:"destination"`
	ETA                  string                  `json:"eta,omitempty"`
	NextExpectedActivity string                  `json:"nextExpectedActivity,omitempty"`
	IsMisdirected        bool                    `json:"isMisdirected"`
	Events               []handlingEventResponse `json:"events"`
}

type handlingEventResponse struct {
	Location       string `json:"location"`
	CompletionTime string `json:"completionTime"`
	Type           string `json:"type"`
	VoyageNumber   string `json:"voyageNumber"`
	IsExpected     bool   `json:"isExpected"`
	Description    string `json:"description"`
}

type locationResponse struct {
	UNLocode string `json:"unLocode"`
	Name     string `json:"name"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// --- Converters ---

const timeFormat = time.RFC3339

func cargoToResponse(c *cargo.Cargo) cargoResponse {
	resp := cargoResponse{
		TrackingID:      string(c.TrackingID),
		Origin:          string(c.Origin),
		Destination:     string(c.RouteSpecification.Destination),
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline.Format(timeFormat),
		Misrouted:       c.Delivery.RoutingStatus == cargo.Misrouted,
		Routed:          c.Delivery.RoutingStatus == cargo.Routed,
	}

	for _, leg := range c.Itinerary.Legs {
		resp.Legs = append(resp.Legs, legToDTO(leg))
	}

	return resp
}

func legToDTO(leg cargo.Leg) legDTO {
	return legDTO{
		VoyageNumber: string(leg.VoyageNumber),
		From:         string(leg.LoadLocation),
		To:           string(leg.UnloadLocation),
		LoadTime:     leg.LoadTime.Format(timeFormat),
		UnloadTime:   leg.UnloadTime.Format(timeFormat),
	}
}

func dtoToItinerary(dto assignRouteRequest) cargo.Itinerary {
	legs := make([]cargo.Leg, len(dto.Legs))
	for i, l := range dto.Legs {
		loadTime, _ := time.Parse(timeFormat, l.LoadTime)
		unloadTime, _ := time.Parse(timeFormat, l.UnloadTime)
		legs[i] = cargo.Leg{
			VoyageNumber:   voyage.Number(l.VoyageNumber),
			LoadLocation:   location.UNLocode(l.From),
			UnloadLocation: location.UNLocode(l.To),
			LoadTime:       loadTime,
			UnloadTime:     unloadTime,
		}
	}
	return cargo.Itinerary{Legs: legs}
}

func locationToResponse(loc *location.Location) locationResponse {
	return locationResponse{
		UNLocode: string(loc.Code),
		Name:     loc.Name,
	}
}
