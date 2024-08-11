package cargo

type RoutingStatus = uint
type TransportStatus = uint

var (
	RoutingStatusNoRouted  RoutingStatus = iota
	RoutingStatusRouted    RoutingStatus
	RoutingStatusMisrouted RoutingStatus
)

var (
	TransportStatusUnknown        TransportStatus = iota
	TransportStatusOnboardCarrier TransportStatus
	TransportStatusInPort         TransportStatus
	TransportStatusNotReceived    TransportStatus
	TransportStatusClaimed        TransportStatus
)

type Delivery struct {
	RoutingStatus   RoutingStatus
	TransportStatus TransportStatus
	// TODO: add the rest
}
