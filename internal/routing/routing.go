package routing

import (
	"context"

	"github.com/roma-glushko/cargo/internal/cargo"
)

// Service finds candidate itineraries for a route specification.
type Service interface {
	FetchRoutes(ctx context.Context, spec cargo.RouteSpecification) ([]cargo.Itinerary, error)
}
