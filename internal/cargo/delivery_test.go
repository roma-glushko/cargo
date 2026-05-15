package cargo_test

import (
	"testing"
	"time"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
	"github.com/stretchr/testify/require"
)

var (
	hongkong  = location.UNLocode("CNHKG")
	newYork   = location.UNLocode("USNYC")
	stockholm = location.UNLocode("SESTO")
	dallas    = location.UNLocode("USDAL")
	hamburg   = location.UNLocode("DEHAM")
	tokyo     = location.UNLocode("JNTKO")

	v100 = voyage.Number("V100")
	v200 = voyage.Number("V200")
	v300 = voyage.Number("V300")

	deadline = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	baseTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
)

func newSpec(origin, dest location.UNLocode) cargo.RouteSpecification {
	return cargo.RouteSpecification{
		Origin:          origin,
		Destination:     dest,
		ArrivalDeadline: deadline,
	}
}

func sampleItinerary() cargo.Itinerary {
	return cargo.Itinerary{
		Legs: []cargo.Leg{
			{VoyageNumber: v100, LoadLocation: hongkong, UnloadLocation: newYork, LoadTime: baseTime, UnloadTime: baseTime.Add(72 * time.Hour)},
			{VoyageNumber: v200, LoadLocation: newYork, UnloadLocation: dallas, LoadTime: baseTime.Add(96 * time.Hour), UnloadTime: baseTime.Add(120 * time.Hour)},
			{VoyageNumber: v300, LoadLocation: dallas, UnloadLocation: stockholm, LoadTime: baseTime.Add(144 * time.Hour), UnloadTime: baseTime.Add(192 * time.Hour)},
		},
	}
}

func event(t cargo.EventType, loc location.UNLocode, vn voyage.Number, completionOffset time.Duration) cargo.HandlingEvent {
	e, _ := cargo.NewHandlingEvent(t, "TEST123", vn, loc, baseTime.Add(completionOffset))
	return e
}

func TestDeriveDelivery_NoEvents(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()

	d := cargo.DeriveDelivery(spec, itin, cargo.HandlingHistory{})

	require.Equal(t, cargo.NotReceived, d.TransportStatus)
	require.Equal(t, cargo.Routed, d.RoutingStatus)
	require.False(t, d.Misdirected)
	require.NotNil(t, d.NextExpectedActivity)
	require.Equal(t, cargo.Receive, d.NextExpectedActivity.Type)
	require.Equal(t, hongkong, d.NextExpectedActivity.Location)
	require.False(t, d.UnloadedAtDestination)
	require.Nil(t, d.LastEvent)
}

func TestDeriveDelivery_NotRouted(t *testing.T) {
	spec := newSpec(hongkong, stockholm)

	d := cargo.DeriveDelivery(spec, cargo.Itinerary{}, cargo.HandlingHistory{})

	require.Equal(t, cargo.NotReceived, d.TransportStatus)
	require.Equal(t, cargo.NotRouted, d.RoutingStatus)
	require.True(t, d.ETA.IsZero())
	require.Nil(t, d.NextExpectedActivity)
}

func TestDeriveDelivery_Received(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()
	history := cargo.HandlingHistory{
		Events: []cargo.HandlingEvent{event(cargo.Receive, hongkong, "", 0)},
	}

	d := cargo.DeriveDelivery(spec, itin, history)

	require.Equal(t, cargo.InPort, d.TransportStatus)
	require.False(t, d.Misdirected)
	require.Equal(t, hongkong, d.LastKnownLocation)
	require.NotNil(t, d.NextExpectedActivity)
	require.Equal(t, cargo.Load, d.NextExpectedActivity.Type)
	require.Equal(t, hongkong, d.NextExpectedActivity.Location)
	require.Equal(t, v100, d.NextExpectedActivity.VoyageNumber)
}

func TestDeriveDelivery_Loaded(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()
	history := cargo.HandlingHistory{
		Events: []cargo.HandlingEvent{
			event(cargo.Receive, hongkong, "", 0),
			event(cargo.Load, hongkong, v100, time.Hour),
		},
	}

	d := cargo.DeriveDelivery(spec, itin, history)

	require.Equal(t, cargo.OnboardCarrier, d.TransportStatus)
	require.Equal(t, v100, d.CurrentVoyage)
	require.False(t, d.Misdirected)
	require.NotNil(t, d.NextExpectedActivity)
	require.Equal(t, cargo.Unload, d.NextExpectedActivity.Type)
	require.Equal(t, newYork, d.NextExpectedActivity.Location)
}

func TestDeriveDelivery_Unloaded_Intermediate(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()
	history := cargo.HandlingHistory{
		Events: []cargo.HandlingEvent{
			event(cargo.Receive, hongkong, "", 0),
			event(cargo.Load, hongkong, v100, time.Hour),
			event(cargo.Unload, newYork, v100, 72*time.Hour),
		},
	}

	d := cargo.DeriveDelivery(spec, itin, history)

	require.Equal(t, cargo.InPort, d.TransportStatus)
	require.Equal(t, newYork, d.LastKnownLocation)
	require.False(t, d.Misdirected)
	require.False(t, d.UnloadedAtDestination)
	require.NotNil(t, d.NextExpectedActivity)
	require.Equal(t, cargo.Load, d.NextExpectedActivity.Type)
	require.Equal(t, newYork, d.NextExpectedActivity.Location)
	require.Equal(t, v200, d.NextExpectedActivity.VoyageNumber)
}

func TestDeriveDelivery_UnloadedAtDestination(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()
	history := cargo.HandlingHistory{
		Events: []cargo.HandlingEvent{
			event(cargo.Unload, stockholm, v300, 192*time.Hour),
		},
	}

	d := cargo.DeriveDelivery(spec, itin, history)

	require.Equal(t, cargo.InPort, d.TransportStatus)
	require.True(t, d.UnloadedAtDestination)
	require.NotNil(t, d.NextExpectedActivity)
	require.Equal(t, cargo.Claim, d.NextExpectedActivity.Type)
	require.Equal(t, stockholm, d.NextExpectedActivity.Location)
}

func TestDeriveDelivery_Claimed(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()
	history := cargo.HandlingHistory{
		Events: []cargo.HandlingEvent{
			event(cargo.Claim, stockholm, "", 200*time.Hour),
		},
	}

	d := cargo.DeriveDelivery(spec, itin, history)

	require.Equal(t, cargo.Claimed, d.TransportStatus)
	require.Nil(t, d.NextExpectedActivity)
}

func TestDeriveDelivery_Misdirected(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()
	history := cargo.HandlingHistory{
		Events: []cargo.HandlingEvent{
			event(cargo.Load, hongkong, v100, time.Hour),
			event(cargo.Unload, tokyo, v100, 48*time.Hour),
		},
	}

	d := cargo.DeriveDelivery(spec, itin, history)

	require.Equal(t, cargo.InPort, d.TransportStatus)
	require.True(t, d.Misdirected)
	require.True(t, d.ETA.IsZero())
	require.Nil(t, d.NextExpectedActivity)
}

func TestDeriveDelivery_Misrouted(t *testing.T) {
	spec := newSpec(hongkong, hamburg)
	itin := sampleItinerary()

	d := cargo.DeriveDelivery(spec, itin, cargo.HandlingHistory{})

	require.Equal(t, cargo.Misrouted, d.RoutingStatus)
	require.True(t, d.ETA.IsZero())
}

func TestDeriveDelivery_ETA_OnTrack(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()

	d := cargo.DeriveDelivery(spec, itin, cargo.HandlingHistory{})

	require.True(t, d.IsOnTrack())
	require.Equal(t, itin.FinalArrivalDate(), d.ETA)
}

func TestDeriveDelivery_Customs(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()
	history := cargo.HandlingHistory{
		Events: []cargo.HandlingEvent{
			event(cargo.Customs, newYork, "", 73*time.Hour),
		},
	}

	d := cargo.DeriveDelivery(spec, itin, history)

	require.Equal(t, cargo.InPort, d.TransportStatus)
	require.False(t, d.Misdirected)
}
