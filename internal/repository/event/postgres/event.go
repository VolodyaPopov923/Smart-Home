package postgres

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/usecase"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		pool: pool,
	}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if event == nil {
		return errors.New("event is nil")
	}

	serialNumber := event.SensorSerialNumber

	if serialNumber == "" {
		var sn string
		err := r.pool.QueryRow(ctx, `SELECT serial_number FROM sensors WHERE id = $1`, event.SensorID).Scan(&sn)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("sensor with id %d not found", event.SensorID)
			}
			return fmt.Errorf("error getting sensor serial number: %w", err)
		}
		serialNumber = sn
	}

	query := `
        INSERT INTO events (timestamp, sensor_serial_number, sensor_id, payload)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.pool.Exec(ctx, query, event.Timestamp, serialNumber, event.SensorID, event.Payload)
	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}
	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	query := `
        SELECT timestamp, sensor_serial_number, sensor_id, payload
        FROM events
        WHERE sensor_id = $1
        ORDER BY timestamp DESC
        LIMIT 1
    `
	var event domain.Event

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&event.Timestamp,
		&event.SensorSerialNumber,
		&event.SensorID,
		&event.Payload,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get last event: %w", err)
	}

	return &event, nil
}
