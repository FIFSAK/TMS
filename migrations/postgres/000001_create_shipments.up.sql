CREATE TABLE IF NOT EXISTS shipments (
    id              VARCHAR(36) PRIMARY KEY,
    reference_number VARCHAR(255) NOT NULL,
    origin          VARCHAR(255) NOT NULL,
    destination     VARCHAR(255) NOT NULL,
    status          VARCHAR(50)  NOT NULL DEFAULT 'pending',
    driver_name     VARCHAR(255) NOT NULL DEFAULT '',
    unit_number     VARCHAR(255) NOT NULL DEFAULT '',
    shipment_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    driver_revenue  DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS shipment_events (
    id          VARCHAR(36) PRIMARY KEY,
    shipment_id VARCHAR(36) NOT NULL REFERENCES shipments(id),
    status      VARCHAR(50) NOT NULL,
    comment     TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shipment_events_shipment_id ON shipment_events(shipment_id);
