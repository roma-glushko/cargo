package cargo_test

import (
	"testing"
	"time"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/stretchr/testify/require"
)

func TestRouteSpecification_IsSatisfiedBy(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	itin := sampleItinerary()

	require.True(t, spec.IsSatisfiedBy(itin))

	wrongDest := newSpec(hongkong, hamburg)
	require.False(t, wrongDest.IsSatisfiedBy(itin))

	pastDeadline := cargo.RouteSpecification{
		Origin:          hongkong,
		Destination:     stockholm,
		ArrivalDeadline: baseTime,
	}
	require.False(t, pastDeadline.IsSatisfiedBy(itin))

	require.False(t, spec.IsSatisfiedBy(cargo.Itinerary{}))
}

func TestItinerary_IsExpected(t *testing.T) {
	itin := sampleItinerary()

	receive := event(cargo.Receive, hongkong, "", 0)
	require.True(t, itin.IsExpected(receive))

	receiveWrong := event(cargo.Receive, newYork, "", 0)
	require.False(t, itin.IsExpected(receiveWrong))

	load := event(cargo.Load, hongkong, v100, time.Hour)
	require.True(t, itin.IsExpected(load))

	loadWrongVoyage := event(cargo.Load, hongkong, v200, time.Hour)
	require.False(t, itin.IsExpected(loadWrongVoyage))

	unload := event(cargo.Unload, newYork, v100, 72*time.Hour)
	require.True(t, itin.IsExpected(unload))

	unloadWrongPlace := event(cargo.Unload, tokyo, v100, 48*time.Hour)
	require.False(t, itin.IsExpected(unloadWrongPlace))

	claim := event(cargo.Claim, stockholm, "", 200*time.Hour)
	require.True(t, itin.IsExpected(claim))

	claimWrong := event(cargo.Claim, newYork, "", 200*time.Hour)
	require.False(t, itin.IsExpected(claimWrong))

	customs := event(cargo.Customs, newYork, "", 73*time.Hour)
	require.True(t, itin.IsExpected(customs))
}

func TestNewHandlingEvent_VoyageConstraints(t *testing.T) {
	_, err := cargo.NewHandlingEvent(cargo.Load, "ABC", v100, hongkong, baseTime)
	require.NoError(t, err)

	_, err = cargo.NewHandlingEvent(cargo.Load, "ABC", "", hongkong, baseTime)
	require.Error(t, err)

	_, err = cargo.NewHandlingEvent(cargo.Receive, "ABC", "", hongkong, baseTime)
	require.NoError(t, err)

	_, err = cargo.NewHandlingEvent(cargo.Receive, "ABC", v100, hongkong, baseTime)
	require.Error(t, err)
}

func TestCargo_Lifecycle(t *testing.T) {
	spec := newSpec(hongkong, stockholm)
	c := cargo.New("ABC123", spec)

	require.Equal(t, cargo.NotReceived, c.Delivery.TransportStatus)
	require.Equal(t, cargo.NotRouted, c.Delivery.RoutingStatus)

	itin := sampleItinerary()
	c.AssignToRoute(itin)

	require.Equal(t, cargo.Routed, c.Delivery.RoutingStatus)
	require.Equal(t, cargo.NotReceived, c.Delivery.TransportStatus)

	newSpec := cargo.RouteSpecification{
		Origin:          hongkong,
		Destination:     hamburg,
		ArrivalDeadline: deadline,
	}
	c.SpecifyNewRoute(newSpec)

	require.Equal(t, cargo.Misrouted, c.Delivery.RoutingStatus)
}
