package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/postgres/sqlcgen"
	"github.com/roma-glushko/cargo/internal/voyage"
)

type VoyageRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

func NewVoyageRepository(pool *pgxpool.Pool) *VoyageRepository {
	return &VoyageRepository{pool: pool, q: sqlcgen.New(pool)}
}

func (r *VoyageRepository) Find(ctx context.Context, number voyage.Number) (*voyage.Voyage, error) {
	_, err := r.q.FindVoyage(ctx, string(number))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, voyage.ErrUnknown
	}
	if err != nil {
		return nil, err
	}

	mRows, err := r.q.FindCarrierMovements(ctx, string(number))
	if err != nil {
		return nil, err
	}

	v := &voyage.Voyage{Number: number}
	for _, m := range mRows {
		v.Schedule.Movements = append(v.Schedule.Movements, voyage.CarrierMovement{
			DepartureLocation: location.UNLocode(m.DepartureLocation),
			ArrivalLocation:   location.UNLocode(m.ArrivalLocation),
			DepartureTime:     m.DepartureTime.Time,
			ArrivalTime:       m.ArrivalTime.Time,
		})
	}

	return v, nil
}

func (r *VoyageRepository) Store(ctx context.Context, v *voyage.Voyage) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	if err := qtx.UpsertVoyage(ctx, string(v.Number)); err != nil {
		return err
	}

	if err := qtx.DeleteCarrierMovements(ctx, string(v.Number)); err != nil {
		return err
	}

	for i, m := range v.Schedule.Movements {
		if err := qtx.InsertCarrierMovement(ctx, sqlcgen.InsertCarrierMovementParams{
			VoyageNumber:      string(v.Number),
			DepartureLocation: string(m.DepartureLocation),
			ArrivalLocation:   string(m.ArrivalLocation),
			DepartureTime:     tstz(m.DepartureTime),
			ArrivalTime:       tstz(m.ArrivalTime),
			Seq:               int32(i),
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
