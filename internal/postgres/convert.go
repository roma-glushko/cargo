package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func tstz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func nullTstz(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func nullText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func nullInt4(v int32, valid bool) pgtype.Int4 {
	return pgtype.Int4{Int32: v, Valid: valid}
}
