# DDD Cargo Tracking System — Complete Specification

A cargo shipping and tracking system built using Domain-Driven Design principles. The system manages the lifecycle of cargo shipments: booking, routing, physical handling at ports and aboard carriers, tracking, and delivery. It demonstrates aggregate boundaries, eventual consistency via asynchronous messaging, and clean separation between domain logic and infrastructure.

## Table of Contents

- [Domain Model Overview](#domain-model-overview)
- [Aggregate 1: Location](#aggregate-1-location)
- [Aggregate 2: Voyage](#aggregate-2-voyage)
- [Aggregate 3: Cargo (Core Aggregate)](#aggregate-3-cargo-core-aggregate)
- [Aggregate 4: Handling Event](#aggregate-4-handling-event)
- [Domain Service: Routing](#domain-service-routing)
- [Application Services (Use Cases)](#application-services-use-cases)
- [Asynchronous Event Pipeline](#asynchronous-event-pipeline)
- [External Integration: Route Finder](#external-integration-route-finder)
- [Interface Layer](#interface-layer)
- [Business Rules Reference](#business-rules-reference)
- [Reference Data](#reference-data)
- [Complete Lifecycle Walkthrough](#complete-lifecycle-walkthrough)

---

## Domain Model Overview

Four aggregates form the domain model. Each aggregate enforces its own invariants and is persisted as a unit.

```
┌─────────────────────────────────────────────────────────────────────┐
│ CARGO (Aggregate Root)                                              │
│  ├── TrackingId (identity, value object)                            │
│  ├── origin → Location (set once, never changes)                    │
│  ├── RouteSpecification (value object, embedded)                    │
│  │    ├── origin → Location                                         │
│  │    ├── destination → Location                                    │
│  │    └── arrivalDeadline: timestamp                                │
│  ├── Itinerary (value object)                                       │
│  │    └── legs: ordered list of Leg (value objects)                  │
│  │         ├── voyage → Voyage                                      │
│  │         ├── loadLocation → Location                              │
│  │         ├── unloadLocation → Location                            │
│  │         ├── loadTime: timestamp                                  │
│  │         └── unloadTime: timestamp                                │
│  └── Delivery (value object, embedded, derived/computed)            │
│       ├── transportStatus: enum                                     │
│       ├── routingStatus: enum                                       │
│       ├── misdirected: boolean                                      │
│       ├── eta: timestamp (nullable)                                 │
│       ├── nextExpectedActivity: HandlingActivity (nullable)         │
│       ├── lastKnownLocation → Location (nullable)                  │
│       ├── currentVoyage → Voyage (nullable)                        │
│       ├── isUnloadedAtDestination: boolean                          │
│       ├── lastEvent → HandlingEvent (nullable)                     │
│       └── calculatedAt: timestamp                                   │
├─────────────────────────────────────────────────────────────────────┤
│ HANDLING EVENT (Aggregate Root)                                     │
│  ├── type: enum (RECEIVE, LOAD, UNLOAD, CLAIM, CUSTOMS)            │
│  ├── cargo → Cargo                                                  │
│  ├── voyage → Voyage (required for LOAD/UNLOAD, null otherwise)    │
│  ├── location → Location                                            │
│  ├── completionTime: timestamp (when it physically happened)        │
│  └── registrationTime: timestamp (when the system recorded it)      │
├─────────────────────────────────────────────────────────────────────┤
│ VOYAGE (Aggregate Root)                                             │
│  ├── VoyageNumber (identity, value object)                          │
│  └── Schedule (value object)                                        │
│       └── carrierMovements: ordered list of CarrierMovement         │
│            ├── departureLocation → Location                         │
│            ├── arrivalLocation → Location                           │
│            ├── departureTime: timestamp                              │
│            └── arrivalTime: timestamp                                │
├─────────────────────────────────────────────────────────────────────┤
│ LOCATION (Aggregate Root)                                           │
│  ├── UnLocode (identity, value object)                              │
│  └── name: string                                                   │
└─────────────────────────────────────────────────────────────────────┘
```

### Base Abstractions

The domain uses three marker interfaces that govern equality semantics:

- **Entity**: compared by identity (e.g., two Cargo objects are equal if they share the same TrackingId, regardless of other fields).
- **Value Object**: compared by all attributes (e.g., two RouteSpecifications are equal if origin, destination, and deadline all match). Immutable.
- **Domain Event**: represents something that happened. Has event identity for deduplication but no mutable lifecycle.

There is also a **Specification** pattern used for composable business rules:

- `isSatisfiedBy(T)` — returns boolean
- Supports `and(spec)`, `or(spec)`, `not(spec)` composition
- RouteSpecification extends this pattern to validate Itineraries

---

## Aggregate 1: Location

Represents a physical shipping location identified by a UN/LOCODE.

### UnLocode (Value Object — Location Identity)

| Field | Type | Constraints |
|-------|------|-------------|
| unlocode | string | Matches pattern `[a-zA-Z]{2}[a-zA-Z2-9]{3}` (2 country chars + 3 location chars). Stored uppercase. |

### Location (Entity — Aggregate Root)

| Field | Type | Constraints |
|-------|------|-------------|
| unLocode | UnLocode | Primary identity. Unique, immutable. |
| name | string | Human-readable name (e.g., "Stockholm"). |

**Null Object**: `Location.UNKNOWN` with code `"XXXXX"` and name `"Unknown location"`. Used as a placeholder when no location is known.

### Location Repository

| Method | Description |
|--------|-------------|
| `find(UnLocode) → Location` | Look up by UN locode |
| `getAll() → List<Location>` | List all shipping locations |
| `store(Location) → Location` | Persist a location |

---

## Aggregate 2: Voyage

Represents a vessel's scheduled journey composed of sequential port-to-port movements.

### VoyageNumber (Value Object — Voyage Identity)

| Field | Type | Constraints |
|-------|------|-------------|
| number | string | Non-null, non-empty. E.g., `"0100S"`, `"V100"`. |

### CarrierMovement (Value Object)

A single leg of a vessel's schedule — departure from one port, arrival at another.

| Field | Type | Constraints |
|-------|------|-------------|
| departureLocation | Location | Where the vessel departs. |
| arrivalLocation | Location | Where the vessel arrives. |
| departureTime | timestamp | Scheduled departure time. |
| arrivalTime | timestamp | Scheduled arrival time. |

**Null Object**: `CarrierMovement.NONE` with `Location.UNKNOWN` for both locations and epoch timestamps.

### Schedule (Value Object)

An ordered collection of CarrierMovements forming a complete voyage schedule.

| Field | Type | Constraints |
|-------|------|-------------|
| carrierMovements | list of CarrierMovement | Immutable, non-empty. |

**Null Object**: `Schedule.EMPTY` — an empty schedule.

### Voyage (Entity — Aggregate Root)

| Field | Type | Constraints |
|-------|------|-------------|
| voyageNumber | VoyageNumber | Primary identity. Unique. |
| schedule | Schedule | The voyage's carrier movements. |

**Null Object**: `Voyage.NONE` with empty voyage number and empty schedule. Used when a handling event has no associated voyage.

**Builder Pattern**: Voyages are constructed using a builder:

```
Builder(voyageNumber, departureLocation)
  .addMovement(arrivalLocation, departureTime, arrivalTime)  // adds movement, next departure = this arrival location
  .addMovement(...)
  .build() → Voyage
```

Each `addMovement` call creates a CarrierMovement from the *previous* arrival location (or the initial departure location) to the given arrival location.

### Voyage Repository

| Method | Description |
|--------|-------------|
| `find(VoyageNumber) → Voyage` | Look up by voyage number |
| `store(Voyage)` | Persist a voyage |

---

## Aggregate 3: Cargo (Core Aggregate)

The central concept. Represents a shipment being tracked through the logistics network.

### TrackingId (Value Object — Cargo Identity)

| Field | Type | Constraints |
|-------|------|-------------|
| id | string | Non-null, non-empty. Generated as UUID by the repository. |

### RouteSpecification (Value Object + Specification)

Captures what the customer wants: move cargo from A to B by a deadline. Acts as a Specification that can validate whether an Itinerary meets the customer's requirements.

| Field | Type | Constraints |
|-------|------|-------------|
| origin | Location | Origin must differ from destination. |
| destination | Location | Where cargo should end up. |
| arrivalDeadline | timestamp | Latest acceptable arrival. |

**Specification logic** — `isSatisfiedBy(Itinerary)` returns true when ALL of:
1. Itinerary's first leg loads at the specification's origin
2. Itinerary's last leg unloads at the specification's destination
3. Itinerary's final arrival time is before the deadline

### Leg (Value Object)

One segment of a planned cargo route — the cargo travels on a specific voyage between two locations.

| Field | Type | Constraints |
|-------|------|-------------|
| voyage | Voyage | The voyage carrying the cargo. |
| loadLocation | Location | Where cargo is loaded onto the voyage. |
| unloadLocation | Location | Where cargo is unloaded from the voyage. |
| loadTime | timestamp | Scheduled load time. |
| unloadTime | timestamp | Scheduled unload time. |

### Itinerary (Value Object)

The complete planned route for a cargo — an ordered list of Legs.

| Field | Type | Constraints |
|-------|------|-------------|
| legs | list of Leg | Non-empty, ordered, no nulls. |

**Null Object**: `Itinerary.EMPTY_ITINERARY` — represents no assigned route.

**Key Methods**:

| Method | Description |
|--------|-------------|
| `initialDepartureLocation()` | First leg's load location (or `Location.UNKNOWN`). |
| `finalArrivalLocation()` | Last leg's unload location (or `Location.UNKNOWN`). |
| `finalArrivalDate()` | Last leg's unload time (or `MAX_TIMESTAMP`). |
| `lastLeg()` | Last leg (or null if empty). |
| `isExpected(HandlingEvent)` | Whether a handling event is consistent with this plan. See rules below. |

**`isExpected(HandlingEvent)` rules:**

| Event Type | Expected when... |
|------------|------------------|
| RECEIVE | Event location equals the first leg's load location |
| LOAD | Any leg has matching load location AND matching voyage |
| UNLOAD | Any leg has matching unload location AND matching voyage |
| CLAIM | Event location equals the last leg's unload location |
| CUSTOMS | Always expected (customs can happen anywhere) |

### HandlingActivity (Value Object)

Describes an expected or observed cargo handling operation. Used inside Delivery to express what should happen next.

| Field | Type | Constraints |
|-------|------|-------------|
| type | HandlingEvent.Type | The type of handling. |
| location | Location | Where the activity happens. |
| voyage | Voyage | Which voyage (nullable — only for LOAD/UNLOAD). |

### TransportStatus (Enum)

Where the cargo physically is in the transport lifecycle.

| Value | Meaning |
|-------|---------|
| `NOT_RECEIVED` | Cargo has been booked but not yet handed to the carrier. |
| `IN_PORT` | Cargo is at a port (received, unloaded, or in customs). |
| `ONBOARD_CARRIER` | Cargo is loaded on a vessel/carrier. |
| `CLAIMED` | Customer has claimed the cargo. Terminal state. |
| `UNKNOWN` | Cannot be determined. |

### RoutingStatus (Enum)

Whether the cargo has a valid planned route.

| Value | Meaning |
|-------|---------|
| `NOT_ROUTED` | No itinerary assigned. |
| `ROUTED` | Assigned itinerary satisfies the route specification. |
| `MISROUTED` | Assigned itinerary does NOT satisfy the route specification (e.g., destination changed). |

### Delivery (Value Object — Computed State)

Represents the current delivery status. This is a **derived value object** — never set directly, always recalculated from the route specification, itinerary, and handling history. It is embedded within the Cargo aggregate.

| Field | Type | Derivation |
|-------|------|------------|
| transportStatus | TransportStatus | Derived from last handling event type. |
| routingStatus | RoutingStatus | Derived from whether itinerary satisfies route specification. |
| misdirected | boolean | True if last event is not expected by itinerary. |
| eta | timestamp? | Final arrival date from itinerary, if cargo is on-track. Null otherwise. |
| nextExpectedActivity | HandlingActivity? | What should happen next. Null when misdirected, claimed, or unrouted. |
| lastKnownLocation | Location? | Location of the last handling event. Null if no events. |
| currentVoyage | Voyage? | The voyage the cargo is currently aboard. Non-null only when transport status is `ONBOARD_CARRIER`. |
| isUnloadedAtDestination | boolean | True if last event is UNLOAD at the route spec destination. |
| lastEvent | HandlingEvent? | The most recent handling event. |
| calculatedAt | timestamp | When this delivery state was computed. |

**Derivation rules in detail:**

**transportStatus** — derived from last event type:

| Last Event Type | Transport Status |
|-----------------|-----------------|
| (none) | `NOT_RECEIVED` |
| `RECEIVE` | `IN_PORT` |
| `LOAD` | `ONBOARD_CARRIER` |
| `UNLOAD` | `IN_PORT` |
| `CUSTOMS` | `IN_PORT` |
| `CLAIM` | `CLAIMED` |

**routingStatus** — derived from itinerary vs route specification:

| Condition | Routing Status |
|-----------|---------------|
| No itinerary (empty) | `NOT_ROUTED` |
| Itinerary satisfies route specification | `ROUTED` |
| Itinerary does NOT satisfy route specification | `MISROUTED` |

**misdirected** — true when a handling event has occurred but the itinerary does not expect it. Calls `itinerary.isExpected(lastEvent)` and inverts.

**eta** — if the cargo is "on track" (routed AND not misdirected), returns the itinerary's final arrival date. Otherwise null.

**isUnloadedAtDestination** — true when last event type is `UNLOAD` AND the event location equals the route specification's destination.

**nextExpectedActivity** — the next thing that should happen to the cargo:

| Current State (last event) | Next Expected Activity |
|---------------------------|----------------------|
| No events yet | `RECEIVE` at route specification origin |
| `RECEIVE` | `LOAD` at first leg's load location on first leg's voyage |
| `LOAD` on leg N | `UNLOAD` at leg N's unload location on leg N's voyage |
| `UNLOAD` at leg N (not final) | `LOAD` at next leg's load location on next leg's voyage |
| `UNLOAD` at final leg | `CLAIM` at last leg's unload location |
| `CLAIM` | No activity (lifecycle complete) |
| Misdirected or unrouted | No activity |

The logic for determining "which leg" after a LOAD: find the leg in the itinerary whose load location and voyage match the event. For UNLOAD: find the leg whose unload location and voyage match, then look at the next leg in sequence.

**on-track** — a cargo is on-track when it is `ROUTED` and not `misdirected`. This is a convenience derivation used to decide whether ETA can be calculated.

**Two update paths for Delivery:**

1. **Synchronous** (within Cargo aggregate): when routing changes (`specifyNewRoute` or `assignToRoute`), Delivery is immediately recalculated using the existing last event.
2. **Asynchronous** (cross-aggregate): when a new handling event is registered for this cargo, `deriveDeliveryProgress(HandlingHistory)` is called later (via message queue), which recalculates Delivery from the full handling history.

### Cargo (Entity — Aggregate Root)

| Field | Type | Constraints |
|-------|------|-------------|
| trackingId | TrackingId | Primary identity. Unique, immutable. |
| origin | Location | Set at creation, never changes (even if route spec changes). |
| routeSpecification | RouteSpecification | Customer requirements. Can be updated. |
| itinerary | Itinerary | Planned route. Can be reassigned. Defaults to `EMPTY_ITINERARY`. |
| delivery | Delivery | Current delivery status. Always derived, never set directly. |

**Key Methods:**

| Method | Behavior |
|--------|----------|
| `specifyNewRoute(routeSpec)` | Updates the route specification. Recalculates delivery synchronously (routing status may become MISROUTED). |
| `assignToRoute(itinerary)` | Assigns a new itinerary. Recalculates delivery synchronously (routing status becomes ROUTED if spec satisfied). |
| `deriveDeliveryProgress(handlingHistory)` | Filters handling history to events for this cargo's tracking ID, then recalculates delivery from the filtered history. Called asynchronously. |

### Cargo Factory

Creates new Cargo aggregates with proper initialization.

**Input**: origin UnLocode, destination UnLocode, arrival deadline.

**Process**:
1. Generate a new TrackingId (UUID-based, from the repository).
2. Look up origin and destination Locations from the repository.
3. Create a RouteSpecification from origin, destination, deadline.
4. Create a new Cargo with the tracking ID and route specification.
5. The new cargo has no itinerary (NOT_ROUTED) and no handling events (NOT_RECEIVED).

### Cargo Repository

| Method | Description |
|--------|-------------|
| `find(TrackingId) → Cargo` | Look up by tracking ID |
| `getAll() → List<Cargo>` | List all cargos |
| `store(Cargo)` | Persist (insert or update) |
| `nextTrackingId() → TrackingId` | Generate a new unique tracking ID (UUID-based) |

---

## Aggregate 4: Handling Event

Represents a real-world cargo handling operation that has occurred — receiving cargo, loading it onto a vessel, unloading, customs inspection, or customer claiming.

### HandlingEvent.Type (Enum)

| Value | Voyage Required? | Description |
|-------|------------------|-------------|
| `RECEIVE` | No (prohibited) | Cargo received at a facility. |
| `LOAD` | Yes (required) | Cargo loaded onto a carrier/vessel. |
| `UNLOAD` | Yes (required) | Cargo unloaded from a carrier/vessel. |
| `CLAIM` | No (prohibited) | Cargo claimed by customer at destination. |
| `CUSTOMS` | No (prohibited) | Cargo going through customs. |

"Required" means the event MUST reference a Voyage. "Prohibited" means the event MUST NOT reference a Voyage.

### HandlingEvent (Domain Event — Aggregate Root)

| Field | Type | Constraints |
|-------|------|-------------|
| type | HandlingEvent.Type | The type of handling. |
| cargo | Cargo | The cargo being handled. |
| voyage | Voyage | The voyage involved. Required for LOAD/UNLOAD, null/NONE for others. |
| location | Location | Where the event occurred. |
| completionTime | timestamp | When the event physically happened. |
| registrationTime | timestamp | When the event was recorded in the system. |

Events are immutable once created. Two constructors enforce the voyage constraint:
- Constructor for LOAD/UNLOAD: requires a non-null Voyage.
- Constructor for RECEIVE/CLAIM/CUSTOMS: no voyage parameter (voyage set to null internally).

**Identity**: Two events are the same if they share the same cargo, voyage, completion time, location, and type.

### HandlingHistory (Value Object)

A collection of handling events for a cargo, providing query methods.

| Field | Type | Constraints |
|-------|------|-------------|
| handlingEvents | collection of HandlingEvent | Can be empty. |

**Null Object**: `HandlingHistory.EMPTY` — empty collection.

**Key Methods:**

| Method | Description |
|--------|-------------|
| `distinctEventsByCompletionTime()` | Returns deduplicated events ordered by completion time. |
| `mostRecentlyCompletedEvent()` | Returns the event with the latest completion time, or null if empty. |
| `filterOnCargo(TrackingId)` | Returns a new HandlingHistory containing only events matching the given tracking ID. |

### Handling Event Factory

Creates HandlingEvent aggregates with validation against existing domain state.

**Input**: registration time, completion time, tracking ID, voyage number (nullable), UN locode, event type.

**Process**:
1. Look up cargo by tracking ID → throws `UnknownCargoException` if not found.
2. Look up voyage by voyage number → throws `UnknownVoyageException` if not found (null voyage number yields `Voyage.NONE`).
3. Look up location by UN locode → throws `UnknownLocationException` if not found.
4. Construct the HandlingEvent with validated domain references.

All three exceptions extend `CannotCreateHandlingEventException`.

### Handling Event Repository

| Method | Description |
|--------|-------------|
| `store(HandlingEvent)` | Persist a handling event. |
| `lookupHandlingHistoryOfCargo(TrackingId) → HandlingHistory` | Retrieve all handling events for a cargo. |

---

## Domain Service: Routing

### RoutingService (Interface)

```
fetchRoutesForSpecification(RouteSpecification) → List<Itinerary>
```

Given a route specification (origin, destination, deadline), returns a list of candidate itineraries that satisfy it. May return an empty list if no viable routes exist.

This is a **domain service** — it encodes domain knowledge about route planning but requires infrastructure support (an external pathfinding algorithm) to execute.

---

## Application Services (Use Cases)

Application services orchestrate domain objects to implement use cases. They are transactional boundaries.

### BookingService

| Method | Use Case | Behavior |
|--------|----------|----------|
| `bookNewCargo(origin, destination, deadline) → TrackingId` | Book a shipment | Uses CargoFactory to create cargo. Persists it. Returns the generated tracking ID. |
| `requestPossibleRoutesForCargo(trackingId) → List<Itinerary>` | Plan a route | Loads cargo, delegates to RoutingService with cargo's route spec. Returns candidate itineraries. |
| `assignCargoToRoute(itinerary, trackingId)` | Assign a route | Loads cargo, calls `cargo.assignToRoute(itinerary)`. Persists. |
| `changeDestination(trackingId, newDestination)` | Re-route cargo | Loads cargo and new location. Creates new RouteSpecification preserving original origin and deadline but with new destination. Calls `cargo.specifyNewRoute(newSpec)`. Persists. |

### HandlingEventService

| Method | Use Case | Behavior |
|--------|----------|----------|
| `registerHandlingEvent(completionTime, trackingId, voyageNumber, unLocode, type)` | Record a handling event | Uses HandlingEventFactory to validate and create event. Persists event. Publishes `cargoWasHandled` application event for async processing. Throws `CannotCreateHandlingEventException` on validation failure. |

### CargoInspectionService

| Method | Use Case | Behavior |
|--------|----------|----------|
| `inspectCargo(trackingId)` | Update cargo delivery status | Loads cargo and its handling history. Calls `cargo.deriveDeliveryProgress(history)`. Persists updated cargo. If cargo is misdirected, publishes `cargoWasMisdirected`. If cargo is unloaded at destination, publishes `cargoHasArrived`. |

### BookingServiceFacade (Anti-Corruption Layer)

A facade that shields the UI/external interface layer from the domain model. Converts between string-based DTOs and domain objects.

| Method | Behavior |
|--------|----------|
| `bookNewCargo(originStr, destStr, deadline) → trackingIdStr` | Wraps BookingService, converts strings to/from value objects. |
| `loadCargoForRouting(trackingIdStr) → CargoRoutingDTO` | Loads cargo, assembles into DTO. |
| `assignCargoToRoute(trackingIdStr, RouteCandidateDTO)` | Converts DTO back to domain Itinerary (looking up Voyages and Locations), delegates to BookingService. |
| `changeDestination(trackingIdStr, destStr)` | Wraps BookingService. |
| `requestPossibleRoutesForCargo(trackingIdStr) → List<RouteCandidateDTO>` | Gets domain itineraries, assembles into DTOs. |
| `listShippingLocations() → List<LocationDTO>` | Lists all locations as DTOs. |
| `listAllCargos() → List<CargoRoutingDTO>` | Lists all cargos as DTOs. |

---

## Asynchronous Event Pipeline

The system uses message queues to decouple handling event registration from cargo state updates. This implements eventual consistency between the HandlingEvent and Cargo aggregates.

### Event Flow

```
                           ┌──────────────────┐
  External Report ────────►│ HandlingReport    │
  (REST API or file)       │ Service           │
                           └────────┬─────────┘
                                    │ publish HandlingEventRegistrationAttempt
                                    ▼
                    ┌───────────────────────────────────┐
                    │ HandlingEventRegistrationAttempt   │
                    │ Queue                              │
                    └───────────────┬───────────────────┘
                                    │ consume
                                    ▼
                           ┌──────────────────┐
                           │ HandlingEvent     │
                           │ Service           │──── validate + persist event
                           └────────┬─────────┘
                                    │ publish cargoWasHandled(trackingId)
                                    ▼
                    ┌───────────────────────────────────┐
                    │ CargoHandled Queue                 │
                    └───────────────┬───────────────────┘
                                    │ consume
                                    ▼
                           ┌──────────────────┐
                           │ CargoInspection   │
                           │ Service           │──── update delivery state
                           └────────┬─────────┘
                                    │ if misdirected or arrived
                                    ▼
                    ┌───────────────────────────────────┐
                    │ MisdirectedCargo / DeliveredCargo  │
                    │ Queue                              │
                    └───────────────────────────────────┘
                                    │ consume
                                    ▼
                              (logging / notification)
```

### Application Events Interface

| Event | Payload | Trigger |
|-------|---------|---------|
| `receivedHandlingEventRegistrationAttempt` | HandlingEventRegistrationAttempt (completionTime, trackingId, voyageNumber, unLocode, type, registrationTime) | REST API or file upload receives a handling report |
| `cargoWasHandled` | TrackingId (as text) | A handling event was successfully registered |
| `cargoWasMisdirected` | TrackingId (as text) | Cargo inspection detected the cargo is off-plan |
| `cargoHasArrived` | TrackingId (as text) | Cargo was unloaded at its final destination |

### Message Queues

| Queue Name | Message Type | Consumer |
|------------|-------------|----------|
| HandlingEventRegistrationAttemptQueue | Serialized HandlingEventRegistrationAttempt | HandlingEventRegistrationAttemptConsumer → HandlingEventService |
| CargoHandledQueue | Text (tracking ID) | CargoHandledConsumer → CargoInspectionService |
| MisdirectedCargoQueue | Text (tracking ID) | Logging consumer |
| DeliveredCargoQueue | Text (tracking ID) | Logging consumer |
| RejectedRegistrationAttemptsQueue | (for failed attempts) | Logging consumer |

---

## External Integration: Route Finder

The routing service delegates to an external pathfinding library via an adapter.

### GraphTraversalService (External API)

```
findShortestPath(originCode, destinationCode, limitations) → List<TransitPath>
```

- `originCode` / `destinationCode`: string location codes (UN locodes)
- `limitations`: key-value properties, notably `DEADLINE` (timestamp as string)
- Returns: list of `TransitPath` objects

### TransitPath / TransitEdge (External Data Structures)

**TransitPath**: ordered list of TransitEdge objects.

**TransitEdge**:

| Field | Type | Description |
|-------|------|-------------|
| edge | string | Voyage identifier (e.g., `"0100S"`) |
| fromNode | string | Departure location code |
| toNode | string | Arrival location code |
| fromDate | timestamp | Departure time |
| toDate | timestamp | Arrival time |

### Algorithm (Reference Implementation)

The reference implementation uses a **randomized path generator** (not a true shortest-path algorithm):

1. Get all location nodes, remove origin and destination.
2. Generate 3–5 candidate paths.
3. For each candidate: pick a random subset of 1–5 intermediate nodes, shuffle them.
4. Build edges: origin → node₁ → node₂ → ... → nodeₙ → destination.
5. For each edge, assign a randomly selected voyage ID and compute times by advancing ~1 day ± random variance per leg.

### Adapter: ExternalRoutingService

Bridges between the domain's `RoutingService` interface and the external pathfinder:

1. Extracts origin/destination codes and deadline from RouteSpecification.
2. Calls GraphTraversalService.
3. Converts each TransitPath → domain Itinerary:
   - TransitEdge.edge → look up Voyage by VoyageNumber
   - TransitEdge.fromNode → look up Location by UnLocode
   - TransitEdge.toNode → look up Location by UnLocode
   - Dates pass through directly.
4. Filters resulting itineraries through `RouteSpecification.isSatisfiedBy()` to drop invalid routes.

---

## Interface Layer

### REST API Endpoints

**Register Handling Events:**

```
POST /handlingReport
Content-Type: application/json

{
  "completionTime": "2024-03-15T10:30:00",
  "trackingIds": ["ABC123"],          // supports multiple cargos in one report
  "type": "LOAD",                     // RECEIVE | LOAD | UNLOAD | CLAIM | CUSTOMS
  "unLocode": "CNSHA",
  "voyageNumber": "0100S"             // required for LOAD/UNLOAD, omit for others
}

Response: 201 Created
```

One report can reference multiple tracking IDs. The parser creates one HandlingEventRegistrationAttempt per tracking ID, each published to the queue independently.

**Track Cargo:**

```
GET /api/track/{trackingId}
Content-Type: application/json

Response: 200 OK
{
  "trackingId": "ABC123",
  "statusText": "In port at Hong Kong",       // human-readable
  "destination": "Helsinki",
  "eta": "2024-03-20T00:00:00Z",
  "nextExpectedActivity": "Load cargo onto voyage 0200T in New York",
  "isMisdirected": false,
  "handlingEvents": [
    {
      "location": "Hong Kong",
      "completionTime": "2024-03-01T00:00:00Z",
      "type": "RECEIVE",
      "voyageNumber": "",
      "isExpected": true,
      "description": "Received in Hong Kong"
    }
  ]
}
```

### Web UI Endpoints (Server-Side Rendered)

**Admin / Booking:**

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/admin/registration` | GET | Show booking form with location list |
| `/admin/register` | POST | Book new cargo (origin, destination, deadline) → redirect to show |
| `/admin/list` | GET | List all cargos with routing status |
| `/admin/show?trackingId=X` | GET | Show cargo details and routing info |
| `/admin/selectItinerary?trackingId=X` | GET | Show candidate routes for assignment |
| `/admin/assignItinerary` | POST | Assign selected route to cargo |
| `/admin/pickNewDestination?trackingId=X` | GET | Show destination change form |
| `/admin/changeDestination` | POST | Change cargo destination |

**Tracking:**

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/track` | GET | Show tracking form |
| `/track` | POST | Track cargo by ID, show status + event history |

### File Upload Interface

A scheduled task (every 5 seconds) scans an upload directory for handling event files.

**File format** (tab/space-delimited, one event per line):

```
2009-03-06 12:30    ABC123    0200T    USNYC    LOAD
2009-03-08 04:00    ABC123    0200T    USDAL    UNLOAD
```

Fields: completionTime, trackingId, voyageNumber, unLocode, eventType.

Successfully processed files are deleted. Lines that fail parsing are written to a rejection file in a separate failure directory.

### DTOs

**CargoRoutingDTO**: trackingId, origin (unlocode), finalDestination (unlocode), arrivalDeadline, misrouted (boolean), legs (list of LegDTO).

**RouteCandidateDTO**: legs (list of LegDTO).

**LegDTO**: voyageNumber, from (unlocode), to (unlocode), loadTime, unloadTime.

**LocationDTO**: unLocode, portName.

---

## Business Rules Reference

### Cargo Lifecycle Rules

1. **TrackingId is immutable** — assigned at booking, never changes.
2. **Origin is immutable** — set at booking from the initial RouteSpecification, never changes even when destination changes.
3. **Route specification can change** — customer may change destination mid-transport. When changed, the cargo's delivery state is recalculated synchronously. If the existing itinerary doesn't satisfy the new spec, routing status becomes `MISROUTED`.
4. **Itinerary can be reassigned** — assigning a new itinerary recalculates delivery synchronously. If the new itinerary satisfies the route spec, routing status becomes `ROUTED`.
5. **Delivery is never set directly** — always derived from (routeSpec, itinerary, handlingHistory).
6. **Misdirection detection** — after any handling event, if the itinerary does not expect that event, the cargo is marked misdirected. This triggers a misdirection notification.
7. **ETA is only available when on-track** — if cargo is misdirected or unrouted, ETA is null.
8. **Lifecycle ends at CLAIM** — once claimed, no further expected activities.

### Handling Event Rules

9. **LOAD and UNLOAD require a voyage** — cannot create these events without specifying which vessel.
10. **RECEIVE, CLAIM, and CUSTOMS prohibit a voyage** — these events happen at facilities, not on vessels.
11. **Events are immutable** — once registered, they cannot be changed.
12. **Completion vs registration time** — completionTime is when the event physically happened; registrationTime is when the system learned about it. They can differ significantly (e.g., a port reports a LOAD hours after it happened).

### Routing Rules

13. **Itinerary must match spec** — first leg starts at origin, last leg ends at destination, final arrival before deadline.
14. **Multiple candidate routes** — the routing service may return several alternatives; the user chooses one.
15. **Route filtering** — candidates that don't satisfy the route spec are discarded by the adapter.

### Consistency Rules

16. **Cargo ↔ Handling Event: eventual consistency** — handling events are registered in their own transaction. Cargo delivery state is updated asynchronously in a subsequent transaction via the message queue.
17. **Cargo ↔ RouteSpec/Itinerary: immediate consistency** — changes to routing are reflected in delivery state within the same transaction.

---

## Reference Data

### Locations

| UN Locode | Name |
|-----------|------|
| CNHKG | Hongkong |
| AUMEL | Melbourne |
| SESTO | Stockholm |
| FIHEL | Helsinki |
| USCHI | Chicago |
| JNTKO | Tokyo |
| DEHAM | Hamburg |
| CNSHA | Shanghai |
| NLRTM | Rotterdam |
| SEGOT | Göteborg |
| CNHGH | Hangzhou |
| USNYC | New York |
| USDAL | Dallas |

### Voyages

**HONGKONG_TO_NEW_YORK (0100S)**:
- Hongkong → Hangzhou
- Hangzhou → Tokyo
- Tokyo → Melbourne
- Melbourne → New York

**NEW_YORK_TO_DALLAS (0200T)**:
- New York → Chicago
- Chicago → Dallas

**DALLAS_TO_HELSINKI (0300A)**:
- Dallas → Hamburg
- Hamburg → Stockholm
- Stockholm → Helsinki

**DALLAS_TO_HELSINKI_ALT (0301S)**:
- Dallas → Helsinki (direct)

**HELSINKI_TO_HONGKONG (0400S)**:
- Helsinki → Rotterdam
- Rotterdam → Shanghai
- Shanghai → Hongkong

**Test-only voyages (V100–V400)** used in scenario tests:

- **V100**: Hongkong → Tokyo → New York
- **V200**: Tokyo → New York → Chicago → Stockholm
- **V300**: Tokyo → Rotterdam → Hamburg → Melbourne → Tokyo
- **V400**: Hamburg → Stockholm → Helsinki → Hamburg

### Pre-loaded Sample Cargos

**Cargo ABC123**:
- Route: Hongkong → Helsinki (deadline: 2009-03-15)
- Itinerary: Hongkong →(0100S)→ New York →(0200T)→ Dallas →(0300A)→ Helsinki
- Events: RECEIVE at Hongkong, LOAD at Hongkong on 0100S, UNLOAD at New York on 0100S
- State: IN_PORT at New York, not misdirected

**Cargo JKL567**:
- Route: Hangzhou → Stockholm (deadline: 2009-03-18)
- Itinerary: Hangzhou →(0100S)→ New York →(0200T)→ Dallas →(0300A)→ Stockholm
- Events: RECEIVE at Hangzhou, LOAD at Hangzhou on 0100S, UNLOAD at New York on 0100S, LOAD at New York on 0100S (wrong voyage — should be 0200T)
- State: ONBOARD_CARRIER on 0100S, **misdirected** (loaded onto wrong voyage)

---

## Complete Lifecycle Walkthrough

This walkthrough covers the full lifecycle of a cargo from Hong Kong to Stockholm, including a misdirection and re-routing. It matches the scenario test and exercises every major business rule.

### Phase 1: Booking

**Action**: Book cargo from Hong Kong to Stockholm, deadline 2009-03-18.

**State after booking**:
- Transport: `NOT_RECEIVED`
- Routing: `NOT_ROUTED`
- Misdirected: false
- ETA: null
- Next expected: null

### Phase 2: Routing

**Action**: Request routes → system returns candidates → assign itinerary: HK →(V100)→ NYC →(V200)→ Chicago →(V200)→ Stockholm.

**State after routing**:
- Transport: `NOT_RECEIVED`
- Routing: `ROUTED`
- ETA: itinerary's final arrival date
- Next expected: `RECEIVE at Hongkong`

### Phase 3: Receive at Hong Kong

**Action**: Register RECEIVE event at Hong Kong.

**State**:
- Transport: `IN_PORT`
- Last location: Hong Kong
- Next expected: `LOAD at Hongkong on V100`

### Phase 4: Load onto V100

**Action**: Register LOAD event at Hong Kong on voyage V100.

**State**:
- Transport: `ONBOARD_CARRIER`
- Current voyage: V100
- Misdirected: false
- Next expected: `UNLOAD at New York on V100`

### Phase 5: Misdirection — Unload at Tokyo (wrong port)

**Action**: Register UNLOAD event at Tokyo on voyage V100.

The itinerary expected unload at New York, not Tokyo.

**State**:
- Transport: `IN_PORT`
- Last location: Tokyo
- **Misdirected: true**
- Next expected: null (can't determine — off plan)
- System publishes `cargoWasMisdirected` event

### Phase 6: Re-routing

**Action**: Specify new route from Tokyo to Stockholm (same deadline). Request new routes. Assign new itinerary: Tokyo →(V300)→ Hamburg →(V400)→ Stockholm.

**Intermediate state** (after specifying new route, before assigning itinerary):
- Routing: `MISROUTED` (old itinerary doesn't satisfy new spec)

**State after assigning new itinerary**:
- Routing: `ROUTED`
- Misdirected: recalculated (now the UNLOAD at Tokyo needs to be re-evaluated against new itinerary)

### Phase 7: Resume Journey

**Actions in sequence**:

1. LOAD at Tokyo on V300 → `ONBOARD_CARRIER`, next: UNLOAD at Hamburg on V300
2. UNLOAD at Hamburg on V300 → `IN_PORT` at Hamburg, next: LOAD at Hamburg on V400
3. LOAD at Hamburg on V400 → `ONBOARD_CARRIER`, next: UNLOAD at Stockholm on V400
4. UNLOAD at Stockholm on V400 → `IN_PORT` at Stockholm, next: CLAIM at Stockholm
   - `isUnloadedAtDestination`: true
   - System publishes `cargoHasArrived` event

### Phase 8: Claim

**Action**: Register CLAIM event at Stockholm.

**Final state**:
- Transport: `CLAIMED`
- Last location: Stockholm
- Misdirected: false
- Next expected: null (lifecycle complete)

---

## Architectural Notes for Reimplementation

1. **Aggregate boundaries matter**: Cargo and HandlingEvent are separate aggregates connected only by TrackingId. They are updated in separate transactions with eventual consistency via messaging.

2. **Delivery is a projection**: The Delivery value object is essentially a materialized view of the cargo's current state. It could alternatively be implemented as a read model / projection rebuilt from events.

3. **The RoutingService is a domain service, not an application service**: It encodes domain knowledge (what makes a valid route) but needs infrastructure support. The adapter pattern bridges to external pathfinding.

4. **Factory validation is a domain concern**: HandlingEventFactory validates that referenced entities exist before allowing event creation. This is not just input validation — it enforces domain invariants.

5. **The Specification pattern** is used for RouteSpecification but could be extended. It provides composable boolean logic (`and`, `or`, `not`) for business rules.

6. **Message ordering is not required**: The system handles events by their completion time, not their arrival order. The "most recently completed event" drives delivery state, so out-of-order processing is safe.

7. **Idempotency**: HandlingHistory deduplicates events by identity (cargo + voyage + time + location + type), making reprocessing safe.

8. **The pathfinder is deliberately external**: It represents a bounded context boundary. In a real system, this would be a separate service with its own data model. The adapter translates between the two models.
