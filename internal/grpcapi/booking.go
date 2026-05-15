package grpcapi

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cargov1 "github.com/roma-glushko/cargo/gen/cargo/v1"
	"github.com/roma-glushko/cargo/internal/booking"
	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
)

type BookingServer struct {
	cargov1.UnimplementedBookingServiceServer
	service   *booking.Service
	cargos    cargo.Repository
	locations location.Repository
}

func (s *BookingServer) BookCargo(ctx context.Context, req *BookCargoRequest) (*BookCargoResponse, error) {
	if req.ArrivalDeadline == nil {
		return nil, status.Error(codes.InvalidArgument, "arrival_deadline is required")
	}

	id, err := s.service.BookNewCargo(
		ctx,
		location.UNLocode(req.Origin),
		location.UNLocode(req.Destination),
		req.ArrivalDeadline.AsTime(),
	)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &BookCargoResponse{TrackingId: string(id)}, nil
}

func (s *BookingServer) ListCargos(ctx context.Context, _ *ListCargosRequest) (*ListCargosResponse, error) {
	cargos, err := s.cargos.FindAll(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &ListCargosResponse{
		Cargos: make([]*Cargo, len(cargos)),
	}
	for i, c := range cargos {
		resp.Cargos[i] = cargoToProto(c)
	}

	return resp, nil
}

func (s *BookingServer) GetCargo(ctx context.Context, req *GetCargoRequest) (*GetCargoResponse, error) {
	c, err := s.cargos.Find(ctx, cargo.TrackingID(req.TrackingId))
	if err != nil {
		return nil, status.Error(codes.NotFound, "cargo not found")
	}

	return &GetCargoResponse{Cargo: cargoToProto(c)}, nil
}

func (s *BookingServer) RequestRoutes(ctx context.Context, req *RequestRoutesRequest) (*RequestRoutesResponse, error) {
	itineraries, err := s.service.RequestPossibleRoutes(ctx, cargo.TrackingID(req.TrackingId))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	resp := &RequestRoutesResponse{
		Itineraries: make([]*Itinerary, len(itineraries)),
	}
	for i, itin := range itineraries {
		legs := make([]*Leg, len(itin.Legs))
		for j, leg := range itin.Legs {
			legs[j] = legToProto(leg)
		}
		resp.Itineraries[i] = &Itinerary{Legs: legs}
	}

	return resp, nil
}

func (s *BookingServer) AssignRoute(ctx context.Context, req *AssignRouteRequest) (*AssignRouteResponse, error) {
	itin := protoToItinerary(req.Legs)
	if err := s.service.AssignToRoute(ctx, cargo.TrackingID(req.TrackingId), itin); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &AssignRouteResponse{}, nil
}

func (s *BookingServer) ChangeDestination(ctx context.Context, req *ChangeDestinationRequest) (*ChangeDestinationResponse, error) {
	if err := s.service.ChangeDestination(ctx, cargo.TrackingID(req.TrackingId), location.UNLocode(req.Destination)); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &ChangeDestinationResponse{}, nil
}

func (s *BookingServer) ListLocations(ctx context.Context, _ *ListLocationsRequest) (*ListLocationsResponse, error) {
	locs, err := s.locations.FindAll(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &ListLocationsResponse{
		Locations: make([]*Location, len(locs)),
	}
	for i, loc := range locs {
		resp.Locations[i] = &Location{
			UnLocode: string(loc.Code),
			Name:     loc.Name,
		}
	}

	return resp, nil
}

type (
	BookCargoRequest          = cargov1.BookCargoRequest
	BookCargoResponse         = cargov1.BookCargoResponse
	ListCargosRequest         = cargov1.ListCargosRequest
	ListCargosResponse        = cargov1.ListCargosResponse
	GetCargoRequest           = cargov1.GetCargoRequest
	GetCargoResponse          = cargov1.GetCargoResponse
	RequestRoutesRequest      = cargov1.RequestRoutesRequest
	RequestRoutesResponse     = cargov1.RequestRoutesResponse
	AssignRouteRequest        = cargov1.AssignRouteRequest
	AssignRouteResponse       = cargov1.AssignRouteResponse
	ChangeDestinationRequest  = cargov1.ChangeDestinationRequest
	ChangeDestinationResponse = cargov1.ChangeDestinationResponse
	ListLocationsRequest      = cargov1.ListLocationsRequest
	ListLocationsResponse     = cargov1.ListLocationsResponse
	Cargo                     = cargov1.Cargo
	Leg                       = cargov1.Leg
	Location                  = cargov1.Location
	Itinerary                 = cargov1.Itinerary
)
