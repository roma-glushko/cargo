-- name: InsertHandlingEvent :exec
INSERT INTO handling_events (type, cargo_id, voyage_number, location, completion_time, registration_time)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: FindHandlingEventsByCargo :many
SELECT type, cargo_id, voyage_number, location, completion_time, registration_time
FROM handling_events WHERE cargo_id = $1 ORDER BY completion_time;
