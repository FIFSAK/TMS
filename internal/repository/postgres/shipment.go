package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FIFSAK/TMS/internal/domain/shipment"
	"github.com/FIFSAK/TMS/pkg/store"
)

type ShipmentRepository struct {
	db *pgxpool.Pool
}

func NewShipmentRepository(db *pgxpool.Pool) *ShipmentRepository {
	return &ShipmentRepository{db: db}
}

func (r *ShipmentRepository) Create(ctx context.Context, s shipment.Shipment) (string, error) {
	id := uuid.New().String()

	query := `INSERT INTO shipments (id, reference_number, origin, destination, status, driver_name, unit_number, shipment_amount, driver_revenue, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(ctx, query,
		id,
		s.ReferenceNumber,
		s.Origin,
		s.Destination,
		string(s.Status),
		s.DriverName,
		s.UnitNumber,
		s.ShipmentAmount,
		s.DriverRevenue,
		s.CreatedAt,
		s.UpdatedAt,
	)
	if err != nil {
		return "", fmt.Errorf("insert shipment: %w", err)
	}

	return id, nil
}

func (r *ShipmentRepository) GetByID(ctx context.Context, id string) (shipment.Shipment, error) {
	query := `SELECT id, reference_number, origin, destination, status, driver_name, unit_number, shipment_amount, driver_revenue, created_at, updated_at
		FROM shipments WHERE id = $1`

	var s shipment.Shipment
	var status string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.ReferenceNumber,
		&s.Origin,
		&s.Destination,
		&status,
		&s.DriverName,
		&s.UnitNumber,
		&s.ShipmentAmount,
		&s.DriverRevenue,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return shipment.Shipment{}, store.ErrorNotFound
		}
		return shipment.Shipment{}, fmt.Errorf("get shipment: %w", err)
	}

	s.Status = shipment.Status(status)
	return s, nil
}

func (r *ShipmentRepository) List(ctx context.Context) ([]shipment.Shipment, error) {
	query := `SELECT id, reference_number, origin, destination, status, driver_name, unit_number, shipment_amount, driver_revenue, created_at, updated_at
		FROM shipments ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list shipments: %w", err)
	}
	defer rows.Close()

	var shipments []shipment.Shipment
	for rows.Next() {
		var s shipment.Shipment
		var status string
		if err := rows.Scan(
			&s.ID,
			&s.ReferenceNumber,
			&s.Origin,
			&s.Destination,
			&status,
			&s.DriverName,
			&s.UnitNumber,
			&s.ShipmentAmount,
			&s.DriverRevenue,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan shipment: %w", err)
		}
		s.Status = shipment.Status(status)
		shipments = append(shipments, s)
	}

	return shipments, nil
}

func (r *ShipmentRepository) UpdateStatus(ctx context.Context, id string, status shipment.Status) error {
	query := `UPDATE shipments SET status = $1, updated_at = $2 WHERE id = $3`

	tag, err := r.db.Exec(ctx, query, string(status), time.Now(), id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return store.ErrorNotFound
	}

	return nil
}

func (r *ShipmentRepository) AddEvent(ctx context.Context, event shipment.Event) (string, error) {
	id := uuid.New().String()

	query := `INSERT INTO shipment_events (id, shipment_id, status, comment, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query,
		id,
		event.ShipmentID,
		string(event.Status),
		event.Comment,
		event.CreatedAt,
	)
	if err != nil {
		return "", fmt.Errorf("insert event: %w", err)
	}

	return id, nil
}

func (r *ShipmentRepository) ListEvents(ctx context.Context, shipmentID string) ([]shipment.Event, error) {
	query := `SELECT id, shipment_id, status, comment, created_at
		FROM shipment_events WHERE shipment_id = $1 ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []shipment.Event
	for rows.Next() {
		var e shipment.Event
		var status string
		if err := rows.Scan(&e.ID, &e.ShipmentID, &status, &e.Comment, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		e.Status = shipment.Status(status)
		events = append(events, e)
	}

	return events, nil
}
