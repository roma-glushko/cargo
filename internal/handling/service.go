package handling

import (
	"context"
	"time"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/eventbus"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

type Service struct {
	events    cargo.HandlingEventRepository
	cargos    cargo.Repository
	locations location.Repository
	voyages   voyage.Repository
	bus       *eventbus.Bus
}

func NewService(
	events cargo.HandlingEventRepository,
	cargos cargo.Repository,
	locations location.Repository,
	voyages voyage.Repository,
	bus *eventbus.Bus,
) *Service {
	return &Service{
		events:    events,
		cargos:    cargos,
		locations: locations,
		voyages:   voyages,
		bus:       bus,
	}
}

func (s *Service) RegisterHandlingEvent(
	ctx context.Context,
	completionTime time.Time,
	trackingID cargo.TrackingID,
	voyageNumber voyage.Number,
	loc location.UNLocode,
	eventType cargo.EventType,
) error {
	if _, err := s.cargos.Find(ctx, trackingID); err != nil {
		return err
	}

	if eventType.RequiresVoyage() {
		if _, err := s.voyages.Find(ctx, voyageNumber); err != nil {
			return err
		}
	}

	if _, err := s.locations.Find(ctx, loc); err != nil {
		return err
	}

	event, err := cargo.NewHandlingEvent(eventType, trackingID, voyageNumber, loc, completionTime)
	if err != nil {
		return err
	}

	if err := s.events.Store(ctx, event); err != nil {
		return err
	}

	s.bus.Publish(eventbus.TopicCargoHandled, trackingID)

	return nil
}
