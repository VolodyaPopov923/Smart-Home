package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"sync"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	users    map[int64]*domain.User
	userLock sync.RWMutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[int64]*domain.User),
	}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if user == nil {
		return errors.New("user cannot be nil")
	}

	r.userLock.Lock()
	defer r.userLock.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if user.ID == 0 {
		maxID := int64(0)
		for id := range r.users {
			if id > maxID {
				maxID = id
			}
		}
		user.ID = maxID + 1
	}
	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.userLock.RLock()
	defer r.userLock.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}
