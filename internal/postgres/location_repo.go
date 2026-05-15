package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/postgres/sqlcgen"
)

type LocationRepository struct {
	q *sqlcgen.Queries
}

func NewLocationRepository(pool *pgxpool.Pool) *LocationRepository {
	return &LocationRepository{q: sqlcgen.New(pool)}
}

func (r *LocationRepository) Find(ctx context.Context, code location.UNLocode) (*location.Location, error) {
	row, err := r.q.FindLocation(ctx, string(code))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, location.ErrUnknown
	}
	if err != nil {
		return nil, err
	}

	return &location.Location{
		Code: location.UNLocode(row.Unlocode),
		Name: row.Name,
	}, nil
}

func (r *LocationRepository) FindAll(ctx context.Context) ([]*location.Location, error) {
	rows, err := r.q.FindAllLocations(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*location.Location, len(rows))
	for i, row := range rows {
		result[i] = &location.Location{
			Code: location.UNLocode(row.Unlocode),
			Name: row.Name,
		}
	}
	return result, nil
}

func (r *LocationRepository) Store(ctx context.Context, loc *location.Location) error {
	return r.q.UpsertLocation(ctx, sqlcgen.UpsertLocationParams{
		Unlocode: string(loc.Code),
		Name:     loc.Name,
	})
}
