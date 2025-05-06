package inmemory

import (
	"context"
	"homework/internal/domain"
	"sync"
)

type SensorOwnerRepository struct {
	data     map[int64]map[int64]domain.SensorOwner
	dataLock sync.RWMutex
}

func NewSensorOwnerRepository() *SensorOwnerRepository {
	return &SensorOwnerRepository{
		data: make(map[int64]map[int64]domain.SensorOwner),
	}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.dataLock.Lock()
	defer r.dataLock.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, exists := r.data[sensorOwner.UserID]
	if !exists {
		r.data[sensorOwner.UserID] = make(map[int64]domain.SensorOwner)
	}

	r.data[sensorOwner.UserID][sensorOwner.SensorID] = sensorOwner
	return nil
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.dataLock.RLock()
	defer r.dataLock.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	userSensors, exists := r.data[userID]
	if !exists {
		return []domain.SensorOwner{}, nil
	}

	result := make([]domain.SensorOwner, 0, len(userSensors))
	for _, sensorOwner := range userSensors {
		result = append(result, sensorOwner)
	}
	return result, nil
}
