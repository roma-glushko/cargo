export interface Cargo {
	trackingId: string;
	origin: string;
	destination: string;
	arrivalDeadline: string;
	misrouted: boolean;
	routed: boolean;
	legs?: Leg[];
}

export interface Leg {
	voyageNumber: string;
	from: string;
	to: string;
	loadTime: string;
	unloadTime: string;
}

export interface Itinerary {
	legs: Leg[];
}

export interface TrackingInfo {
	trackingId: string;
	statusText: string;
	destination: string;
	eta?: string;
	nextExpectedActivity?: string;
	isMisdirected: boolean;
	events: HandlingEvent[];
}

export interface HandlingEvent {
	location: string;
	completionTime: string;
	type: string;
	voyageNumber: string;
	isExpected: boolean;
	description: string;
}

export interface Location {
	unLocode: string;
	name: string;
}
