package cargo

import (
	"errors"
	"sort"
	"time"

	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

type HandlingEvent struct {
	Type             EventType
	CargoID          TrackingID
	VoyageNumber     voyage.Number
	Location         location.UNLocode
	CompletionTime   time.Time
	RegistrationTime time.Time
}

func NewHandlingEvent(
	eventType EventType,
	cargoID TrackingID,
	voyageNum voyage.Number,
	loc location.UNLocode,
	completionTime time.Time,
) (HandlingEvent, error) {
	if eventType.RequiresVoyage() && voyageNum == "" {
		return HandlingEvent{}, errors.New("voyage required for LOAD/UNLOAD events")
	}
	if eventType.ProhibitsVoyage() && voyageNum != "" {
		return HandlingEvent{}, errors.New("voyage must not be set for RECEIVE/CLAIM/CUSTOMS events")
	}
	return HandlingEvent{
		Type:             eventType,
		CargoID:          cargoID,
		VoyageNumber:     voyageNum,
		Location:         loc,
		CompletionTime:   completionTime,
		RegistrationTime: time.Now(),
	}, nil
}

type HandlingHistory struct {
	Events []HandlingEvent
}

func (h HandlingHistory) MostRecentEvent() *HandlingEvent {
	if len(h.Events) == 0 {
		return nil
	}

	sorted := make([]HandlingEvent, len(h.Events))
	copy(sorted, h.Events)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CompletionTime.Before(sorted[j].CompletionTime)
	})

	last := sorted[len(sorted)-1]
	return &last
}

func (h HandlingHistory) FilterByCargo(id TrackingID) HandlingHistory {
	var filtered []HandlingEvent
	for _, e := range h.Events {
		if e.CargoID == id {
			filtered = append(filtered, e)
		}
	}
	return HandlingHistory{Events: filtered}
}
