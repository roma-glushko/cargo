package grpcapi

import (
	"github.com/roma-glushko/cargo/internal/booking"
	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/handling"
	"github.com/roma-glushko/cargo/internal/location"

	cargov1 "github.com/roma-glushko/cargo/gen/cargo/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewServer(
	bookingService *booking.Service,
	handlingService *handling.Service,
	cargos cargo.Repository,
	locations location.Repository,
	events cargo.HandlingEventRepository,
) *grpc.Server {
	s := grpc.NewServer()

	cargov1.RegisterBookingServiceServer(s, &BookingServer{
		service:   bookingService,
		cargos:    cargos,
		locations: locations,
	})
	cargov1.RegisterTrackingServiceServer(s, &TrackingServer{
		cargos: cargos,
		events: events,
	})
	cargov1.RegisterHandlingServiceServer(s, &HandlingServer{
		service: handlingService,
	})

	reflection.Register(s)

	return s
}
