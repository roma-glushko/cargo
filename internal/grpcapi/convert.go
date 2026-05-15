package grpcapi

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	cargov1 "github.com/roma-glushko/cargo/gen/cargo/v1"
	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

func cargoToProto(c *cargo.Cargo) *cargov1.Cargo {
	pb := &cargov1.Cargo{
		TrackingId:      string(c.TrackingID),
		Origin:          string(c.Origin),
		Destination:     string(c.RouteSpecification.Destination),
		ArrivalDeadline: timestamppb.New(c.RouteSpecification.ArrivalDeadline),
		Misrouted:       c.Delivery.RoutingStatus == cargo.Misrouted,
		Routed:          c.Delivery.RoutingStatus == cargo.Routed,
	}

	for _, leg := range c.Itinerary.Legs {
		pb.Legs = append(pb.Legs, legToProto(leg))
	}

	return pb
}

func legToProto(leg cargo.Leg) *cargov1.Leg {
	return &cargov1.Leg{
		VoyageNumber: string(leg.VoyageNumber),
		From:         string(leg.LoadLocation),
		To:           string(leg.UnloadLocation),
		LoadTime:     timestamppb.New(leg.LoadTime),
		UnloadTime:   timestamppb.New(leg.UnloadTime),
	}
}

func protoToItinerary(legs []*cargov1.Leg) cargo.Itinerary {
	result := make([]cargo.Leg, len(legs))

	for i, l := range legs {
		result[i] = cargo.Leg{
			VoyageNumber:   voyage.Number(l.VoyageNumber),
			LoadLocation:   location.UNLocode(l.From),
			UnloadLocation: location.UNLocode(l.To),
		}

		if l.LoadTime != nil {
			result[i].LoadTime = l.LoadTime.AsTime()
		}

		if l.UnloadTime != nil {
			result[i].UnloadTime = l.UnloadTime.AsTime()
		}
	}

	return cargo.Itinerary{Legs: result}
}
