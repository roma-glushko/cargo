package cargo

import (
	"context"
	"errors"
	"time"

	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

var ErrUnknown = errors.New("unknown cargo")

type TrackingID string

type RouteSpecification struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

func (s RouteSpecification) IsSatisfiedBy(itin Itinerary) bool {
	if itin.IsEmpty() {
		return false
	}

	return itin.InitialDepartureLocation() == s.Origin &&
		itin.FinalArrivalLocation() == s.Destination &&
		itin.FinalArrivalDate().Before(s.ArrivalDeadline)
}

type Leg struct {
	VoyageNumber   voyage.Number
	LoadLocation   location.UNLocode
	UnloadLocation location.UNLocode
	LoadTime       time.Time
	UnloadTime     time.Time
}

type Itinerary struct {
	Legs []Leg
}

func (i Itinerary) IsEmpty() bool {
	return len(i.Legs) == 0
}

func (i Itinerary) InitialDepartureLocation() location.UNLocode {
	if i.IsEmpty() {
		return ""
	}

	return i.Legs[0].LoadLocation
}

func (i Itinerary) FinalArrivalLocation() location.UNLocode {
	if i.IsEmpty() {
		return ""
	}

	return i.Legs[len(i.Legs)-1].UnloadLocation
}

func (i Itinerary) FinalArrivalDate() time.Time {
	if i.IsEmpty() {
		return time.Time{}
	}

	return i.Legs[len(i.Legs)-1].UnloadTime
}

func (i Itinerary) IsExpected(event HandlingEvent) bool {
	if i.IsEmpty() {
		return true
	}

	switch event.Type {
	case Receive:
		return i.Legs[0].LoadLocation == event.Location

	case Load:
		return i.findLegByLoadLocation(event.Location, event.VoyageNumber) != nil

	case Unload:
		for _, leg := range i.Legs {
			if leg.UnloadLocation == event.Location && leg.VoyageNumber == event.VoyageNumber {
				return true
			}
		}
		return false

	case Claim:
		return i.Legs[len(i.Legs)-1].UnloadLocation == event.Location

	case Customs:
		return true

	default:
		return true
	}
}

func (i Itinerary) findLegByLoadLocation(loc location.UNLocode, vn voyage.Number) *Leg {
	for idx := range i.Legs {
		if i.Legs[idx].LoadLocation == loc && i.Legs[idx].VoyageNumber == vn {
			return &i.Legs[idx]
		}
	}
	return nil
}

type HandlingActivity struct {
	Type         EventType
	Location     location.UNLocode
	VoyageNumber voyage.Number
}

type Cargo struct {
	TrackingID         TrackingID
	Origin             location.UNLocode
	RouteSpecification RouteSpecification
	Itinerary          Itinerary
	Delivery           Delivery
}

func New(id TrackingID, spec RouteSpecification) *Cargo {
	c := &Cargo{
		TrackingID:         id,
		Origin:             spec.Origin,
		RouteSpecification: spec,
	}
	c.Delivery = DeriveDelivery(spec, c.Itinerary, HandlingHistory{})
	return c
}

func (c *Cargo) SpecifyNewRoute(spec RouteSpecification) {
	c.RouteSpecification = spec
	c.Delivery = DeriveDelivery(spec, c.Itinerary, HandlingHistory{Events: c.lastEventAsSlice()})
}

func (c *Cargo) AssignToRoute(itin Itinerary) {
	c.Itinerary = itin
	c.Delivery = DeriveDelivery(c.RouteSpecification, itin, HandlingHistory{Events: c.lastEventAsSlice()})
}

func (c *Cargo) DeriveDeliveryProgress(history HandlingHistory) {
	filtered := history.FilterByCargo(c.TrackingID)
	c.Delivery = DeriveDelivery(c.RouteSpecification, c.Itinerary, filtered)
}

func (c *Cargo) lastEventAsSlice() []HandlingEvent {
	if c.Delivery.LastEvent == nil {
		return nil
	}
	return []HandlingEvent{*c.Delivery.LastEvent}
}

type Repository interface {
	Find(ctx context.Context, id TrackingID) (*Cargo, error)
	FindAll(ctx context.Context) ([]*Cargo, error)
	Store(ctx context.Context, c *Cargo) error
	NextTrackingID() TrackingID
}

type HandlingEventRepository interface {
	Store(ctx context.Context, event HandlingEvent) error
	LookupHandlingHistory(ctx context.Context, id TrackingID) (HandlingHistory, error)
}
