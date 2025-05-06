package postgres

import (
	"context"
	"fmt"
	"homework/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SensorOwnerRepository struct {
	pool *pgxpool.Pool
}

func NewSensorOwnerRepository(pool *pgxpool.Pool) *SensorOwnerRepository {
	return &SensorOwnerRepository{
		pool: pool,
	}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	var count int
	checkQuery := `
        SELECT COUNT(*) FROM sensors_users 
        WHERE sensor_id = $1 AND user_id = $2
    `
	err := r.pool.QueryRow(ctx, checkQuery, sensorOwner.SensorID, sensorOwner.UserID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing sensor owner: %w", err)
	}

	if count > 0 {
		return nil
	}

	query := `
        INSERT INTO sensors_users (sensor_id, user_id)
        VALUES ($1, $2)
    `
	_, err = r.pool.Exec(ctx, query, sensorOwner.SensorID, sensorOwner.UserID)
	if err != nil {
		return fmt.Errorf("failed to save sensor owner: %w", err)
	}
	return nil
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	query := `
        SELECT sensor_id, user_id
        FROM sensors_users
        WHERE user_id = $1
    `
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sensor owners: %w", err)
	}
	defer rows.Close()

	var result []domain.SensorOwner
	for rows.Next() {
		var so domain.SensorOwner
		if err := rows.Scan(&so.SensorID, &so.UserID); err != nil {
			return nil, fmt.Errorf("failed to scan sensor owner: %w", err)
		}
		result = append(result, so)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through sensor owners: %w", err)
	}
	return result, nil
}
