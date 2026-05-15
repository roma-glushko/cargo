package inspection

import (
	"context"
	"log/slog"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/eventbus"
)

type Service struct {
	cargos cargo.Repository
	events cargo.HandlingEventRepository
	bus    *eventbus.Bus
	logger *slog.Logger
}

func NewService(
	cargos cargo.Repository,
	events cargo.HandlingEventRepository,
	bus *eventbus.Bus,
	logger *slog.Logger,
) *Service {
	return &Service{
		cargos: cargos,
		events: events,
		bus:    bus,
		logger: logger,
	}
}

func (s *Service) InspectCargo(id cargo.TrackingID) {
	ctx := context.Background()

	c, err := s.cargos.Find(ctx, id)
	if err != nil {
		s.logger.Error("failed to find cargo for inspection", "trackingId", string(id), "error", err)
		return
	}

	history, err := s.events.LookupHandlingHistory(ctx, id)
	if err != nil {
		s.logger.Error("failed to lookup handling history", "trackingId", string(id), "error", err)
		return
	}

	c.DeriveDeliveryProgress(history)

	if err := s.cargos.Store(ctx, c); err != nil {
		s.logger.Error("failed to store updated cargo", "trackingId", string(id), "error", err)
		return
	}

	if c.Delivery.Misdirected {
		s.bus.Publish(eventbus.TopicCargoMisdirected, id)
	}

	if c.Delivery.UnloadedAtDestination {
		s.bus.Publish(eventbus.TopicCargoArrived, id)
	}
}
