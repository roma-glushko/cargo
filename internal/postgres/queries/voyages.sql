-- name: FindVoyage :one
SELECT number FROM voyages WHERE number = $1;

-- name: UpsertVoyage :exec
INSERT INTO voyages (number) VALUES ($1) ON CONFLICT (number) DO NOTHING;

-- name: DeleteCarrierMovements :exec
DELETE FROM carrier_movements WHERE voyage_number = $1;

-- name: InsertCarrierMovement :exec
INSERT INTO carrier_movements
    (voyage_number, departure_location, arrival_location, departure_time, arrival_time, seq)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: FindCarrierMovements :many
SELECT departure_location, arrival_location, departure_time, arrival_time
FROM carrier_movements WHERE voyage_number = $1 ORDER BY seq;
