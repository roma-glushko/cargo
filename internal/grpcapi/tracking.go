package grpcapi

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	cargov1 "github.com/roma-glushko/cargo/gen/cargo/v1"
	"github.com/roma-glushko/cargo/internal/cargo"
)

type TrackingServer struct {
	cargov1.UnimplementedTrackingServiceServer
	cargos cargo.Repository
	events cargo.HandlingEventRepository
}

func (s *TrackingServer) TrackCargo(ctx context.Context, req *cargov1.TrackCargoRequest) (*cargov1.TrackCargoResponse, error) {
	id := cargo.TrackingID(req.TrackingId)

	c, err := s.cargos.Find(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "cargo not found")
	}

	history, err := s.events.LookupHandlingHistory(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &cargov1.TrackCargoResponse{
		TrackingId:    string(c.TrackingID),
		StatusText:    formatStatusText(c),
		Destination:   string(c.RouteSpecification.Destination),
		IsMisdirected: c.Delivery.Misdirected,
	}

	if !c.Delivery.ETA.IsZero() {
		resp.Eta = timestamppb.New(c.Delivery.ETA)
	}

	if act := c.Delivery.NextExpectedActivity; act != nil {
		desc := formatActivity(act)
		resp.NextExpectedActivity = &desc
	}

	for _, e := range history.Events {
		resp.Events = append(resp.Events, &cargov1.HandlingEvent{
			Location:       string(e.Location),
			CompletionTime: timestamppb.New(e.CompletionTime),
			Type:           e.Type.String(),
			VoyageNumber:   string(e.VoyageNumber),
			IsExpected:     c.Itinerary.IsExpected(e),
			Description:    formatEventDescription(e),
		})
	}

	return resp, nil
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
