package voyage

import (
	"context"
	"errors"
	"time"

	"github.com/roma-glushko/cargo/internal/location"
)

var ErrUnknown = errors.New("unknown voyage")

type Number string

type CarrierMovement struct {
	DepartureLocation location.UNLocode
	ArrivalLocation   location.UNLocode
	DepartureTime     time.Time
	ArrivalTime       time.Time
}

type Schedule struct {
	Movements []CarrierMovement
}

type Voyage struct {
	Number   Number
	Schedule Schedule
}

type Repository interface {
	Find(ctx context.Context, number Number) (*Voyage, error)
	Store(ctx context.Context, v *Voyage) error
}

// Builder constructs a Voyage by chaining movements from a starting location.
type Builder struct {
	number    Number
	current   location.UNLocode
	movements []CarrierMovement
}

func NewBuilder(number Number, departureLocation location.UNLocode) *Builder {
	return &Builder{
		number:  number,
		current: departureLocation,
	}
}

func (b *Builder) AddMovement(arrival location.UNLocode, departureTime, arrivalTime time.Time) *Builder {
	b.movements = append(b.movements, CarrierMovement{
		DepartureLocation: b.current,
		ArrivalLocation:   arrival,
		DepartureTime:     departureTime,
		ArrivalTime:       arrivalTime,
	})
	b.current = arrival
	return b
}

func (b *Builder) Build() *Voyage {
	return &Voyage{
		Number: b.number,
		Schedule: Schedule{
			Movements: b.movements,
		},
	}
}
