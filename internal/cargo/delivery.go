package cargo

import (
	"time"

	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

type RoutingStatus int

const (
	NotRouted RoutingStatus = iota
	Routed
	Misrouted
)

func (s RoutingStatus) String() string {
	switch s {
	case NotRouted:
		return "NOT_ROUTED"
	case Routed:
		return "ROUTED"
	case Misrouted:
		return "MISROUTED"
	default:
		return "UNKNOWN"
	}
}

type TransportStatus int

const (
	NotReceived TransportStatus = iota
	InPort
	OnboardCarrier
	Claimed
)

func (s TransportStatus) String() string {
	switch s {
	case NotReceived:
		return "NOT_RECEIVED"
	case InPort:
		return "IN_PORT"
	case OnboardCarrier:
		return "ONBOARD_CARRIER"
	case Claimed:
		return "CLAIMED"
	default:
		return "UNKNOWN"
	}
}

type EventType int

const (
	Receive EventType = iota
	Load
	Unload
	Claim
	Customs
)

func (t EventType) String() string {
	switch t {
	case Receive:
		return "RECEIVE"
	case Load:
		return "LOAD"
	case Unload:
		return "UNLOAD"
	case Claim:
		return "CLAIM"
	case Customs:
		return "CUSTOMS"
	default:
		return "UNKNOWN"
	}
}

func ParseEventType(s string) (EventType, bool) {
	switch s {
	case "RECEIVE":
		return Receive, true
	case "LOAD":
		return Load, true
	case "UNLOAD":
		return Unload, true
	case "CLAIM":
		return Claim, true
	case "CUSTOMS":
		return Customs, true
	default:
		return 0, false
	}
}

func (t EventType) RequiresVoyage() bool {
	return t == Load || t == Unload
}

func (t EventType) ProhibitsVoyage() bool {
	return !t.RequiresVoyage()
}

type Delivery struct {
	TransportStatus       TransportStatus
	RoutingStatus         RoutingStatus
	Misdirected           bool
	ETA                   time.Time
	NextExpectedActivity  *HandlingActivity
	LastKnownLocation     location.UNLocode
	CurrentVoyage         voyage.Number
	UnloadedAtDestination bool
	LastEvent             *HandlingEvent
	CalculatedAt          time.Time
}

func (d Delivery) IsOnTrack() bool {
	return d.RoutingStatus == Routed && !d.Misdirected
}

// DeriveDelivery computes the current delivery state from the route specification,
// itinerary, and handling history. This is a pure function — no side effects.
func DeriveDelivery(spec RouteSpecification, itin Itinerary, history HandlingHistory) Delivery {
	lastEvent := history.MostRecentEvent()

	transportStatus := deriveTransportStatus(lastEvent)
	routingStatus := deriveRoutingStatus(spec, itin)
	misdirected := deriveMisdirected(lastEvent, itin)
	onTrack := routingStatus == Routed && !misdirected

	var eta time.Time
	if onTrack {
		eta = itin.FinalArrivalDate()
	}

	return Delivery{
		TransportStatus:       transportStatus,
		RoutingStatus:         routingStatus,
		Misdirected:           misdirected,
		ETA:                   eta,
		NextExpectedActivity:  deriveNextExpectedActivity(lastEvent, itin, spec, routingStatus, misdirected),
		LastKnownLocation:     deriveLastKnownLocation(lastEvent),
		CurrentVoyage:         deriveCurrentVoyage(transportStatus, lastEvent),
		UnloadedAtDestination: deriveUnloadedAtDestination(lastEvent, spec),
		LastEvent:             lastEvent,
		CalculatedAt:          time.Now(),
	}
}

func deriveTransportStatus(lastEvent *HandlingEvent) TransportStatus {
	if lastEvent == nil {
		return NotReceived
	}
	switch lastEvent.Type {
	case Receive:
		return InPort
	case Load:
		return OnboardCarrier
	case Unload:
		return InPort
	case Customs:
		return InPort
	case Claim:
		return Claimed
	default:
		return NotReceived
	}
}

func deriveRoutingStatus(spec RouteSpecification, itin Itinerary) RoutingStatus {
	if itin.IsEmpty() {
		return NotRouted
	}
	if spec.IsSatisfiedBy(itin) {
		return Routed
	}
	return Misrouted
}

func deriveMisdirected(lastEvent *HandlingEvent, itin Itinerary) bool {
	if lastEvent == nil {
		return false
	}
	return !itin.IsExpected(*lastEvent)
}

func deriveLastKnownLocation(lastEvent *HandlingEvent) location.UNLocode {
	if lastEvent == nil {
		return ""
	}
	return lastEvent.Location
}

func deriveCurrentVoyage(ts TransportStatus, lastEvent *HandlingEvent) voyage.Number {
	if ts == OnboardCarrier && lastEvent != nil {
		return lastEvent.VoyageNumber
	}
	return ""
}

func deriveUnloadedAtDestination(lastEvent *HandlingEvent, spec RouteSpecification) bool {
	if lastEvent == nil {
		return false
	}
	return lastEvent.Type == Unload && lastEvent.Location == spec.Destination
}

func deriveNextExpectedActivity(
	lastEvent *HandlingEvent,
	itin Itinerary,
	spec RouteSpecification,
	routingStatus RoutingStatus,
	misdirected bool,
) *HandlingActivity {
	if routingStatus != Routed || misdirected {
		return nil
	}

	if lastEvent == nil {
		return &HandlingActivity{
			Type:     Receive,
			Location: spec.Origin,
		}
	}

	switch lastEvent.Type {
	case Receive:
		if itin.IsEmpty() {
			return nil
		}
		firstLeg := itin.Legs[0]
		return &HandlingActivity{
			Type:         Load,
			Location:     firstLeg.LoadLocation,
			VoyageNumber: firstLeg.VoyageNumber,
		}

	case Load:
		leg := itin.findLegByLoadLocation(lastEvent.Location, lastEvent.VoyageNumber)
		if leg == nil {
			return nil
		}
		return &HandlingActivity{
			Type:         Unload,
			Location:     leg.UnloadLocation,
			VoyageNumber: leg.VoyageNumber,
		}

	case Unload:
		for i, leg := range itin.Legs {
			if leg.UnloadLocation == lastEvent.Location && leg.VoyageNumber == lastEvent.VoyageNumber {
				if i+1 < len(itin.Legs) {
					nextLeg := itin.Legs[i+1]
					return &HandlingActivity{
						Type:         Load,
						Location:     nextLeg.LoadLocation,
						VoyageNumber: nextLeg.VoyageNumber,
					}
				}
				return &HandlingActivity{
					Type:     Claim,
					Location: leg.UnloadLocation,
				}
			}
		}
		return nil

	case Claim:
		return nil

	default:
		return nil
	}
}
