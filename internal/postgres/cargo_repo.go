package postgres

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/postgres/sqlcgen"
	"github.com/roma-glushko/cargo/internal/voyage"
)

type CargoRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

func NewCargoRepository(pool *pgxpool.Pool) *CargoRepository {
	return &CargoRepository{pool: pool, q: sqlcgen.New(pool)}
}

func (r *CargoRepository) Find(ctx context.Context, id cargo.TrackingID) (*cargo.Cargo, error) {
	row, err := r.q.FindCargo(ctx, string(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, cargo.ErrUnknown
	}
	if err != nil {
		return nil, err
	}

	c := cargoFromRow(row)

	legs, err := r.q.FindLegs(ctx, string(id))
	if err != nil {
		return nil, err
	}
	c.Itinerary = cargo.Itinerary{Legs: legsFromRows(legs)}

	return c, nil
}

func (r *CargoRepository) FindAll(ctx context.Context) ([]*cargo.Cargo, error) {
	rows, err := r.q.FindAllCargos(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*cargo.Cargo, len(rows))
	for i, row := range rows {
		c := cargoFromRow(row)

		legs, err := r.q.FindLegs(ctx, string(c.TrackingID))
		if err != nil {
			return nil, err
		}
		c.Itinerary = cargo.Itinerary{Legs: legsFromRows(legs)}

		result[i] = c
	}

	return result, nil
}

func (r *CargoRepository) Store(ctx context.Context, c *cargo.Cargo) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	var nextActivityType nullableInt4
	var nextActivityLocation, nextActivityVoyage string
	if c.Delivery.NextExpectedActivity != nil {
		nextActivityType = nullableInt4{int32(c.Delivery.NextExpectedActivity.Type), true}
		nextActivityLocation = string(c.Delivery.NextExpectedActivity.Location)
		nextActivityVoyage = string(c.Delivery.NextExpectedActivity.VoyageNumber)
	}

	var lastEventType nullableInt4
	var lastEventCargo, lastEventVoyage, lastEventLocation string
	var lastEventCompletion, lastEventRegistration nullableTime
	if c.Delivery.LastEvent != nil {
		e := c.Delivery.LastEvent
		lastEventType = nullableInt4{int32(e.Type), true}
		lastEventCargo = string(e.CargoID)
		lastEventVoyage = string(e.VoyageNumber)
		lastEventLocation = string(e.Location)
		lastEventCompletion = nullableTime{e.CompletionTime, true}
		lastEventRegistration = nullableTime{e.RegistrationTime, true}
	}

	if err := qtx.UpsertCargo(ctx, sqlcgen.UpsertCargoParams{
		TrackingID:            string(c.TrackingID),
		Origin:                string(c.Origin),
		SpecOrigin:            string(c.RouteSpecification.Origin),
		SpecDestination:       string(c.RouteSpecification.Destination),
		SpecDeadline:          tstz(c.RouteSpecification.ArrivalDeadline),
		TransportStatus:       int32(c.Delivery.TransportStatus),
		RoutingStatus:         int32(c.Delivery.RoutingStatus),
		Misdirected:           c.Delivery.Misdirected,
		Eta:                   nullTstz(c.Delivery.ETA),
		NextActivityType:      nullInt4(nextActivityType.val, nextActivityType.valid),
		NextActivityLocation:  nullText(nextActivityLocation),
		NextActivityVoyage:    nullText(nextActivityVoyage),
		LastLocation:          nullText(string(c.Delivery.LastKnownLocation)),
		CurrentVoyage:         nullText(string(c.Delivery.CurrentVoyage)),
		UnloadedAtDest:        c.Delivery.UnloadedAtDestination,
		LastEventType:         nullInt4(lastEventType.val, lastEventType.valid),
		LastEventCargo:        nullText(lastEventCargo),
		LastEventVoyage:       nullText(lastEventVoyage),
		LastEventLocation:     nullText(lastEventLocation),
		LastEventCompletion:   nullableTstz(lastEventCompletion),
		LastEventRegistration: nullableTstz(lastEventRegistration),
		CalculatedAt:          nullTstz(c.Delivery.CalculatedAt),
	}); err != nil {
		return err
	}

	if err := qtx.DeleteLegs(ctx, string(c.TrackingID)); err != nil {
		return err
	}

	for i, leg := range c.Itinerary.Legs {
		if err := qtx.InsertLeg(ctx, sqlcgen.InsertLegParams{
			CargoID:        string(c.TrackingID),
			VoyageNumber:   string(leg.VoyageNumber),
			LoadLocation:   string(leg.LoadLocation),
			UnloadLocation: string(leg.UnloadLocation),
			LoadTime:       tstz(leg.LoadTime),
			UnloadTime:     tstz(leg.UnloadTime),
			Seq:            int32(i),
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *CargoRepository) NextTrackingID() cargo.TrackingID {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return cargo.TrackingID(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]))
}

type nullableInt4 struct {
	val   int32
	valid bool
}

type nullableTime struct {
	t     time.Time
	valid bool
}

func nullableTstz(nt nullableTime) pgtype.Timestamptz {
	if !nt.valid {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: nt.t, Valid: true}
}

func cargoFromRow(row sqlcgen.Cargo) *cargo.Cargo {
	c := &cargo.Cargo{
		TrackingID: cargo.TrackingID(row.TrackingID),
		Origin:     location.UNLocode(row.Origin),
		RouteSpecification: cargo.RouteSpecification{
			Origin:          location.UNLocode(row.SpecOrigin),
			Destination:     location.UNLocode(row.SpecDestination),
			ArrivalDeadline: row.SpecDeadline.Time,
		},
		Delivery: cargo.Delivery{
			TransportStatus:       cargo.TransportStatus(row.TransportStatus),
			RoutingStatus:         cargo.RoutingStatus(row.RoutingStatus),
			Misdirected:           row.Misdirected,
			UnloadedAtDestination: row.UnloadedAtDest,
		},
	}

	if row.Eta.Valid {
		c.Delivery.ETA = row.Eta.Time
	}
	if row.LastLocation.Valid {
		c.Delivery.LastKnownLocation = location.UNLocode(row.LastLocation.String)
	}
	if row.CurrentVoyage.Valid {
		c.Delivery.CurrentVoyage = voyage.Number(row.CurrentVoyage.String)
	}
	if row.CalculatedAt.Valid {
		c.Delivery.CalculatedAt = row.CalculatedAt.Time
	}

	if row.NextActivityType.Valid {
		act := &cargo.HandlingActivity{
			Type: cargo.EventType(row.NextActivityType.Int32),
		}
		if row.NextActivityLocation.Valid {
			act.Location = location.UNLocode(row.NextActivityLocation.String)
		}
		if row.NextActivityVoyage.Valid {
			act.VoyageNumber = voyage.Number(row.NextActivityVoyage.String)
		}
		c.Delivery.NextExpectedActivity = act
	}

	if row.LastEventType.Valid {
		evt := &cargo.HandlingEvent{
			Type: cargo.EventType(row.LastEventType.Int32),
		}
		if row.LastEventCargo.Valid {
			evt.CargoID = cargo.TrackingID(row.LastEventCargo.String)
		}
		if row.LastEventVoyage.Valid {
			evt.VoyageNumber = voyage.Number(row.LastEventVoyage.String)
		}
		if row.LastEventLocation.Valid {
			evt.Location = location.UNLocode(row.LastEventLocation.String)
		}
		if row.LastEventCompletion.Valid {
			evt.CompletionTime = row.LastEventCompletion.Time
		}
		if row.LastEventRegistration.Valid {
			evt.RegistrationTime = row.LastEventRegistration.Time
		}
		c.Delivery.LastEvent = evt
	}

	return c
}

func legsFromRows(rows []sqlcgen.FindLegsRow) []cargo.Leg {
	legs := make([]cargo.Leg, len(rows))
	for i, row := range rows {
		legs[i] = cargo.Leg{
			VoyageNumber:   voyage.Number(row.VoyageNumber),
			LoadLocation:   location.UNLocode(row.LoadLocation),
			UnloadLocation: location.UNLocode(row.UnloadLocation),
			LoadTime:       row.LoadTime.Time,
			UnloadTime:     row.UnloadTime.Time,
		}
	}
	return legs
}
