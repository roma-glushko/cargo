CREATE TABLE locations (
    unlocode VARCHAR(5) PRIMARY KEY,
    name     VARCHAR(100) NOT NULL
);

CREATE TABLE voyages (
    number VARCHAR(10) PRIMARY KEY
);

CREATE TABLE carrier_movements (
    id                  SERIAL PRIMARY KEY,
    voyage_number       VARCHAR(10) NOT NULL REFERENCES voyages(number),
    departure_location  VARCHAR(5) NOT NULL REFERENCES locations(unlocode),
    arrival_location    VARCHAR(5) NOT NULL REFERENCES locations(unlocode),
    departure_time      TIMESTAMPTZ NOT NULL,
    arrival_time        TIMESTAMPTZ NOT NULL,
    seq                 INT NOT NULL,
    UNIQUE (voyage_number, seq)
);

CREATE TABLE cargos (
    tracking_id            VARCHAR(36) PRIMARY KEY,
    origin                 VARCHAR(5) NOT NULL REFERENCES locations(unlocode),
    spec_origin            VARCHAR(5) NOT NULL REFERENCES locations(unlocode),
    spec_destination       VARCHAR(5) NOT NULL REFERENCES locations(unlocode),
    spec_deadline          TIMESTAMPTZ NOT NULL,
    transport_status       INT NOT NULL DEFAULT 0,
    routing_status         INT NOT NULL DEFAULT 0,
    misdirected            BOOLEAN NOT NULL DEFAULT FALSE,
    eta                    TIMESTAMPTZ,
    next_activity_type     INT,
    next_activity_location VARCHAR(5),
    next_activity_voyage   VARCHAR(10),
    last_location          VARCHAR(5),
    current_voyage         VARCHAR(10),
    unloaded_at_dest       BOOLEAN NOT NULL DEFAULT FALSE,
    last_event_type        INT,
    last_event_cargo       VARCHAR(36),
    last_event_voyage      VARCHAR(10),
    last_event_location    VARCHAR(5),
    last_event_completion  TIMESTAMPTZ,
    last_event_registration TIMESTAMPTZ,
    calculated_at          TIMESTAMPTZ
);

CREATE TABLE legs (
    id              SERIAL PRIMARY KEY,
    cargo_id        VARCHAR(36) NOT NULL REFERENCES cargos(tracking_id) ON DELETE CASCADE,
    voyage_number   VARCHAR(10) NOT NULL,
    load_location   VARCHAR(5) NOT NULL,
    unload_location VARCHAR(5) NOT NULL,
    load_time       TIMESTAMPTZ NOT NULL,
    unload_time     TIMESTAMPTZ NOT NULL,
    seq             INT NOT NULL,
    UNIQUE (cargo_id, seq)
);

CREATE TABLE handling_events (
    id                SERIAL PRIMARY KEY,
    type              INT NOT NULL,
    cargo_id          VARCHAR(36) NOT NULL REFERENCES cargos(tracking_id),
    voyage_number     VARCHAR(10) NOT NULL DEFAULT '',
    location          VARCHAR(5) NOT NULL,
    completion_time   TIMESTAMPTZ NOT NULL,
    registration_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_handling_events_cargo ON handling_events(cargo_id);
CREATE INDEX idx_handling_events_cargo_completion ON handling_events(cargo_id, completion_time);
