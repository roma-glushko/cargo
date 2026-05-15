-- name: FindCargo :one
SELECT tracking_id, origin, spec_origin, spec_destination, spec_deadline,
       transport_status, routing_status, misdirected, eta,
       next_activity_type, next_activity_location, next_activity_voyage,
       last_location, current_voyage, unloaded_at_dest,
       last_event_type, last_event_cargo, last_event_voyage,
       last_event_location, last_event_completion, last_event_registration,
       calculated_at
FROM cargos WHERE tracking_id = $1;

-- name: FindAllCargos :many
SELECT tracking_id, origin, spec_origin, spec_destination, spec_deadline,
       transport_status, routing_status, misdirected, eta,
       next_activity_type, next_activity_location, next_activity_voyage,
       last_location, current_voyage, unloaded_at_dest,
       last_event_type, last_event_cargo, last_event_voyage,
       last_event_location, last_event_completion, last_event_registration,
       calculated_at
FROM cargos ORDER BY tracking_id;

-- name: UpsertCargo :exec
INSERT INTO cargos (
    tracking_id, origin, spec_origin, spec_destination, spec_deadline,
    transport_status, routing_status, misdirected, eta,
    next_activity_type, next_activity_location, next_activity_voyage,
    last_location, current_voyage, unloaded_at_dest,
    last_event_type, last_event_cargo, last_event_voyage,
    last_event_location, last_event_completion, last_event_registration,
    calculated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
ON CONFLICT (tracking_id) DO UPDATE SET
    origin = EXCLUDED.origin,
    spec_origin = EXCLUDED.spec_origin,
    spec_destination = EXCLUDED.spec_destination,
    spec_deadline = EXCLUDED.spec_deadline,
    transport_status = EXCLUDED.transport_status,
    routing_status = EXCLUDED.routing_status,
    misdirected = EXCLUDED.misdirected,
    eta = EXCLUDED.eta,
    next_activity_type = EXCLUDED.next_activity_type,
    next_activity_location = EXCLUDED.next_activity_location,
    next_activity_voyage = EXCLUDED.next_activity_voyage,
    last_location = EXCLUDED.last_location,
    current_voyage = EXCLUDED.current_voyage,
    unloaded_at_dest = EXCLUDED.unloaded_at_dest,
    last_event_type = EXCLUDED.last_event_type,
    last_event_cargo = EXCLUDED.last_event_cargo,
    last_event_voyage = EXCLUDED.last_event_voyage,
    last_event_location = EXCLUDED.last_event_location,
    last_event_completion = EXCLUDED.last_event_completion,
    last_event_registration = EXCLUDED.last_event_registration,
    calculated_at = EXCLUDED.calculated_at;

-- name: DeleteLegs :exec
DELETE FROM legs WHERE cargo_id = $1;

-- name: InsertLeg :exec
INSERT INTO legs (cargo_id, voyage_number, load_location, unload_location, load_time, unload_time, seq)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: FindLegs :many
SELECT voyage_number, load_location, unload_location, load_time, unload_time
FROM legs WHERE cargo_id = $1 ORDER BY seq;
