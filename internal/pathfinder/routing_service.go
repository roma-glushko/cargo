package pathfinder

import (
	"context"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

// RoutingServiceAdapter adapts the external pathfinder to the routing.Service interface.
type RoutingServiceAdapter struct {
	locations location.Repository
	voyages   voyage.Repository
}

func NewRoutingService(locations location.Repository, voyages voyage.Repository) *RoutingServiceAdapter {
	return &RoutingServiceAdapter{
		locations: locations,
		voyages:   voyages,
	}
}

func (a *RoutingServiceAdapter) FetchRoutes(ctx context.Context, spec cargo.RouteSpecification) ([]cargo.Itinerary, error) {
	allLocations, err := a.locations.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var locationCodes []string
	for _, loc := range allLocations {
		locationCodes = append(locationCodes, string(loc.Code))
	}

	voyageNumbers := collectVoyageNumbers(ctx, a.voyages)

	if len(voyageNumbers) == 0 {
		voyageNumbers = []string{"0100S", "0200T", "0300A", "0301S", "0400S"}
	}

	transitPaths := FindShortestPath(
		string(spec.Origin),
		string(spec.Destination),
		locationCodes,
		voyageNumbers,
	)

	var itineraries []cargo.Itinerary
	for _, path := range transitPaths {
		itin := convertToItinerary(path)
		if spec.IsSatisfiedBy(itin) {
			itineraries = append(itineraries, itin)
		}
	}

	return itineraries, nil
}

func convertToItinerary(path TransitPath) cargo.Itinerary {
	legs := make([]cargo.Leg, len(path.Edges))
	for i, edge := range path.Edges {
		legs[i] = cargo.Leg{
			VoyageNumber:   voyage.Number(edge.VoyageNumber),
			LoadLocation:   location.UNLocode(edge.From),
			UnloadLocation: location.UNLocode(edge.To),
			LoadTime:       edge.FromDate,
			UnloadTime:     edge.ToDate,
		}
	}
	return cargo.Itinerary{Legs: legs}
}

func collectVoyageNumbers(ctx context.Context, voyages voyage.Repository) []string {
	knownVoyages := []string{"0100S", "0200T", "0300A", "0301S", "0400S", "V100", "V200", "V300", "V400"}

	var result []string
	for _, vn := range knownVoyages {
		if _, err := voyages.Find(ctx, voyage.Number(vn)); err == nil {
			result = append(result, vn)
		}
	}
	return result
}
