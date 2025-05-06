package postgres

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/usecase"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SensorRepository struct {
	pool *pgxpool.Pool
}

func NewSensorRepository(pool *pgxpool.Pool) *SensorRepository {
	return &SensorRepository{
		pool: pool,
	}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	if sensor == nil {
		return errors.New("sensor is nil")
	}

	if sensor.RegisteredAt.IsZero() {
		sensor.RegisteredAt = time.Now()
	}

	sensor.LastActivity = time.Now()

	query := `
		INSERT INTO sensors (
			serial_number, type, current_state, description, 
			is_active, registered_at, last_activity
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (serial_number) DO UPDATE SET 
			type = EXCLUDED.type,
			current_state = EXCLUDED.current_state,
			description = EXCLUDED.description,
			is_active = EXCLUDED.is_active,
			last_activity = EXCLUDED.last_activity
		RETURNING id, registered_at, last_activity
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		sensor.SerialNumber,
		sensor.Type,
		sensor.CurrentState,
		sensor.Description,
		sensor.IsActive,
		sensor.RegisteredAt,
		sensor.LastActivity,
	).Scan(&sensor.ID, &sensor.RegisteredAt, &sensor.LastActivity)
	if err != nil {
		return fmt.Errorf("failed to upsert sensor: %w", err)
	}

	return nil
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	query := `
		SELECT id, serial_number, type, current_state, description, is_active, registered_at, last_activity
		FROM sensors
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sensors: %w", err)
	}
	defer rows.Close()

	var sensors []domain.Sensor
	for rows.Next() {
		var s domain.Sensor
		if err := rows.Scan(
			&s.ID,
			&s.SerialNumber,
			&s.Type,
			&s.CurrentState,
			&s.Description,
			&s.IsActive,
			&s.RegisteredAt,
			&s.LastActivity,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sensor: %w", err)
		}
		sensors = append(sensors, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through sensors: %w", err)
	}
	return sensors, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	query := `
		SELECT id, serial_number, type, current_state, description, is_active, registered_at, last_activity
		FROM sensors
		WHERE id = $1
	`
	var s domain.Sensor
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.SerialNumber,
		&s.Type,
		&s.CurrentState,
		&s.Description,
		&s.IsActive,
		&s.RegisteredAt,
		&s.LastActivity,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrSensorNotFound
		}
		return nil, fmt.Errorf("failed to get sensor: %w", err)
	}
	return &s, nil
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	query := `
		SELECT id, serial_number, type, current_state, description, is_active, registered_at, last_activity
		FROM sensors
		WHERE serial_number = $1
	`
	var s domain.Sensor
	err := r.pool.QueryRow(ctx, query, sn).Scan(
		&s.ID,
		&s.SerialNumber,
		&s.Type,
		&s.CurrentState,
		&s.Description,
		&s.IsActive,
		&s.RegisteredAt,
		&s.LastActivity,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrSensorNotFound
		}
		return nil, fmt.Errorf("failed to get sensor by serial number: %w", err)
	}
	return &s, nil
}
