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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	query := `
		WITH input_data AS (
			SELECT $1::bigint AS id, $2::text AS name
		)
		INSERT INTO users (id, name)
		SELECT 
			CASE WHEN id.id = 0 THEN nextval('users_id_seq') ELSE id.id END,
			id.name
		FROM input_data id
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query, user.ID, user.Name).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	fmt.Printf("User saved with ID: %d, Name: %s\n", user.ID, user.Name)
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
        SELECT id, name
        FROM users
        WHERE id = $1
    `
	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
