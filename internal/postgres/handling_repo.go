package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/postgres/sqlcgen"
	"github.com/roma-glushko/cargo/internal/voyage"
)

type HandlingEventRepository struct {
	q *sqlcgen.Queries
}

func NewHandlingEventRepository(pool *pgxpool.Pool) *HandlingEventRepository {
	return &HandlingEventRepository{q: sqlcgen.New(pool)}
}

func (r *HandlingEventRepository) Store(ctx context.Context, event cargo.HandlingEvent) error {
	return r.q.InsertHandlingEvent(ctx, sqlcgen.InsertHandlingEventParams{
		Type:             int32(event.Type),
		CargoID:          string(event.CargoID),
		VoyageNumber:     string(event.VoyageNumber),
		Location:         string(event.Location),
		CompletionTime:   tstz(event.CompletionTime),
		RegistrationTime: tstz(event.RegistrationTime),
	})
}

func (r *HandlingEventRepository) LookupHandlingHistory(ctx context.Context, id cargo.TrackingID) (cargo.HandlingHistory, error) {
	rows, err := r.q.FindHandlingEventsByCargo(ctx, string(id))
	if err != nil {
		return cargo.HandlingHistory{}, err
	}

	events := make([]cargo.HandlingEvent, len(rows))
	for i, row := range rows {
		events[i] = cargo.HandlingEvent{
			Type:             cargo.EventType(row.Type),
			CargoID:          cargo.TrackingID(row.CargoID),
			VoyageNumber:     voyage.Number(row.VoyageNumber),
			Location:         location.UNLocode(row.Location),
			CompletionTime:   row.CompletionTime.Time,
			RegistrationTime: row.RegistrationTime.Time,
		}
	}

	return cargo.HandlingHistory{Events: events}, nil
}
