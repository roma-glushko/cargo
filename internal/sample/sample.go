package sample

import (
	"context"
	"time"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

func Populate(
	ctx context.Context,
	locations location.Repository,
	voyages voyage.Repository,
	cargos cargo.Repository,
	events cargo.HandlingEventRepository,
) error {
	for _, loc := range Locations {
		if err := locations.Store(ctx, loc); err != nil {
			return err
		}
	}

	for _, v := range AllVoyages {
		if err := voyages.Store(ctx, v); err != nil {
			return err
		}
	}

	// Only seed sample cargos if they don't exist yet
	if _, err := cargos.Find(ctx, "ABC123"); err != nil {
		if err := populateSampleCargos(ctx, cargos, events); err != nil {
			return err
		}
	}

	return nil
}

func populateSampleCargos(ctx context.Context, cargos cargo.Repository, events cargo.HandlingEventRepository) error {
	abc123 := cargo.New("ABC123", cargo.RouteSpecification{
		Origin:          "CNHKG",
		Destination:     "FIHEL",
		ArrivalDeadline: ts(2009, 3, 15),
	})

	abc123.AssignToRoute(cargo.Itinerary{
		Legs: []cargo.Leg{
			{VoyageNumber: "0100S", LoadLocation: "CNHKG", UnloadLocation: "USNYC", LoadTime: ts(2009, 3, 2), UnloadTime: ts(2009, 3, 5)},
			{VoyageNumber: "0200T", LoadLocation: "USNYC", UnloadLocation: "USDAL", LoadTime: ts(2009, 3, 6), UnloadTime: ts(2009, 3, 8)},
			{VoyageNumber: "0300A", LoadLocation: "USDAL", UnloadLocation: "FIHEL", LoadTime: ts(2009, 3, 9), UnloadTime: ts(2009, 3, 12)},
		},
	})

	if err := cargos.Store(ctx, abc123); err != nil {
		return err
	}

	abc123Events := []cargo.HandlingEvent{
		mustEvent(cargo.Receive, "ABC123", "", "CNHKG", ts(2009, 3, 1)),
		mustEvent(cargo.Load, "ABC123", "0100S", "CNHKG", ts(2009, 3, 2)),
		mustEvent(cargo.Unload, "ABC123", "0100S", "USNYC", ts(2009, 3, 5)),
	}

	for _, e := range abc123Events {
		if err := events.Store(ctx, e); err != nil {
			return err
		}
	}

	history, _ := events.LookupHandlingHistory(ctx, "ABC123")
	abc123.DeriveDeliveryProgress(history)

	if err := cargos.Store(ctx, abc123); err != nil {
		return err
	}

	jkl567 := cargo.New("JKL567", cargo.RouteSpecification{
		Origin:          "CNHGH",
		Destination:     "SESTO",
		ArrivalDeadline: ts(2009, 3, 18),
	})

	jkl567.AssignToRoute(cargo.Itinerary{
		Legs: []cargo.Leg{
			{VoyageNumber: "0100S", LoadLocation: "CNHGH", UnloadLocation: "USNYC", LoadTime: ts(2009, 3, 3), UnloadTime: ts(2009, 3, 5)},
			{VoyageNumber: "0200T", LoadLocation: "USNYC", UnloadLocation: "USDAL", LoadTime: ts(2009, 3, 6), UnloadTime: ts(2009, 3, 8)},
			{VoyageNumber: "0300A", LoadLocation: "USDAL", UnloadLocation: "SESTO", LoadTime: ts(2009, 3, 9), UnloadTime: ts(2009, 3, 12)},
		},
	})

	if err := cargos.Store(ctx, jkl567); err != nil {
		return err
	}

	jkl567Events := []cargo.HandlingEvent{
		mustEvent(cargo.Receive, "JKL567", "", "CNHGH", ts(2009, 3, 1)),
		mustEvent(cargo.Load, "JKL567", "0100S", "CNHGH", ts(2009, 3, 3)),
		mustEvent(cargo.Unload, "JKL567", "0100S", "USNYC", ts(2009, 3, 5)),
		mustEvent(cargo.Load, "JKL567", "0100S", "USNYC", ts(2009, 3, 5)),
	}

	for _, e := range jkl567Events {
		if err := events.Store(ctx, e); err != nil {
			return err
		}
	}

	history, _ = events.LookupHandlingHistory(ctx, "JKL567")
	jkl567.DeriveDeliveryProgress(history)

	return cargos.Store(ctx, jkl567)
}

func mustEvent(eventType cargo.EventType, cargoID cargo.TrackingID, voyageNum string, loc string, completionTime time.Time) cargo.HandlingEvent {
	e, err := cargo.NewHandlingEvent(eventType, cargoID, voyage.Number(voyageNum), location.UNLocode(loc), completionTime)
	if err != nil {
		panic(err)
	}
	return e
}
