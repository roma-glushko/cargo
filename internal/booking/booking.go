package booking

import (
	"context"
	"time"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/routing"
)

type Service struct {
	cargos    cargo.Repository
	locations location.Repository
	routing   routing.Service
}

func NewService(cargos cargo.Repository, locations location.Repository, routing routing.Service) *Service {
	return &Service{
		cargos:    cargos,
		locations: locations,
		routing:   routing,
	}
}

func (s *Service) BookNewCargo(ctx context.Context, origin, destination location.UNLocode, deadline time.Time) (cargo.TrackingID, error) {
	if _, err := s.locations.Find(ctx, origin); err != nil {
		return "", err
	}
	if _, err := s.locations.Find(ctx, destination); err != nil {
		return "", err
	}

	id := s.cargos.NextTrackingID()
	spec := cargo.RouteSpecification{
		Origin:          origin,
		Destination:     destination,
		ArrivalDeadline: deadline,
	}

	c := cargo.New(id, spec)
	if err := s.cargos.Store(ctx, c); err != nil {
		return "", err
	}

	return id, nil
}

func (s *Service) RequestPossibleRoutes(ctx context.Context, id cargo.TrackingID) ([]cargo.Itinerary, error) {
	c, err := s.cargos.Find(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.routing.FetchRoutes(ctx, c.RouteSpecification)
}

func (s *Service) AssignToRoute(ctx context.Context, id cargo.TrackingID, itin cargo.Itinerary) error {
	c, err := s.cargos.Find(ctx, id)
	if err != nil {
		return err
	}

	c.AssignToRoute(itin)
	return s.cargos.Store(ctx, c)
}

func (s *Service) ChangeDestination(ctx context.Context, id cargo.TrackingID, destination location.UNLocode) error {
	c, err := s.cargos.Find(ctx, id)
	if err != nil {
		return err
	}

	if _, err := s.locations.Find(ctx, destination); err != nil {
		return err
	}

	newSpec := cargo.RouteSpecification{
		Origin:          c.RouteSpecification.Origin,
		Destination:     destination,
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
	}
	c.SpecifyNewRoute(newSpec)

	return s.cargos.Store(ctx, c)
}
