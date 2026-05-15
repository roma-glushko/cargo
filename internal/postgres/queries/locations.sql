-- name: FindLocation :one
SELECT unlocode, name FROM locations WHERE unlocode = $1;

-- name: FindAllLocations :many
SELECT unlocode, name FROM locations ORDER BY unlocode;

-- name: UpsertLocation :exec
INSERT INTO locations (unlocode, name) VALUES ($1, $2)
ON CONFLICT (unlocode) DO UPDATE SET name = EXCLUDED.name;
