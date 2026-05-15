package inmem

import (
	"context"
	"crypto/rand"
	"fmt"
	"sort"
	"sync"

	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

// CargoRepository is a thread-safe in-memory cargo store.
type CargoRepository struct {
	mu     sync.RWMutex
	cargos map[cargo.TrackingID]*cargo.Cargo
}

func NewCargoRepository() *CargoRepository {
	return &CargoRepository{
		cargos: make(map[cargo.TrackingID]*cargo.Cargo),
	}
}

func (r *CargoRepository) Find(_ context.Context, id cargo.TrackingID) (*cargo.Cargo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.cargos[id]
	if !ok {
		return nil, cargo.ErrUnknown
	}
	cp := *c
	return &cp, nil
}

func (r *CargoRepository) FindAll(_ context.Context) ([]*cargo.Cargo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*cargo.Cargo, 0, len(r.cargos))
	for _, c := range r.cargos {
		cp := *c
		result = append(result, &cp)
	}
	return result, nil
}

func (r *CargoRepository) Store(_ context.Context, c *cargo.Cargo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *c
	r.cargos[c.TrackingID] = &cp
	return nil
}

func (r *CargoRepository) NextTrackingID() cargo.TrackingID {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return cargo.TrackingID(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]))
}

// LocationRepository is a thread-safe in-memory location store.
type LocationRepository struct {
	mu        sync.RWMutex
	locations map[location.UNLocode]*location.Location
}

func NewLocationRepository() *LocationRepository {
	return &LocationRepository{
		locations: make(map[location.UNLocode]*location.Location),
	}
}

func (r *LocationRepository) Find(_ context.Context, code location.UNLocode) (*location.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	loc, ok := r.locations[code]
	if !ok {
		return nil, location.ErrUnknown
	}
	cp := *loc
	return &cp, nil
}

func (r *LocationRepository) FindAll(_ context.Context) ([]*location.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*location.Location, 0, len(r.locations))
	for _, loc := range r.locations {
		cp := *loc
		result = append(result, &cp)
	}
	return result, nil
}

func (r *LocationRepository) Store(_ context.Context, loc *location.Location) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *loc
	r.locations[loc.Code] = &cp
	return nil
}

// VoyageRepository is a thread-safe in-memory voyage store.
type VoyageRepository struct {
	mu      sync.RWMutex
	voyages map[voyage.Number]*voyage.Voyage
}

func NewVoyageRepository() *VoyageRepository {
	return &VoyageRepository{
		voyages: make(map[voyage.Number]*voyage.Voyage),
	}
}

func (r *VoyageRepository) Find(_ context.Context, number voyage.Number) (*voyage.Voyage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.voyages[number]
	if !ok {
		return nil, voyage.ErrUnknown
	}
	cp := *v
	return &cp, nil
}

func (r *VoyageRepository) Store(_ context.Context, v *voyage.Voyage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *v
	r.voyages[v.Number] = &cp
	return nil
}

// HandlingEventRepository is a thread-safe in-memory handling event store.
type HandlingEventRepository struct {
	mu     sync.RWMutex
	events []cargo.HandlingEvent
}

func NewHandlingEventRepository() *HandlingEventRepository {
	return &HandlingEventRepository{}
}

func (r *HandlingEventRepository) Store(_ context.Context, event cargo.HandlingEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, event)
	return nil
}

func (r *HandlingEventRepository) LookupHandlingHistory(_ context.Context, id cargo.TrackingID) (cargo.HandlingHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []cargo.HandlingEvent
	for _, e := range r.events {
		if e.CargoID == id {
			filtered = append(filtered, e)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CompletionTime.Before(filtered[j].CompletionTime)
	})

	return cargo.HandlingHistory{Events: filtered}, nil
}
