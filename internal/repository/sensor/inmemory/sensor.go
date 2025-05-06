package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

type SensorRepository struct {
	sensors     map[int64]*domain.Sensor
	sensorsBySN map[string]*domain.Sensor
	mu          sync.RWMutex
	lastID      int64
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{
		sensors:     make(map[int64]*domain.Sensor),
		sensorsBySN: make(map[string]*domain.Sensor),
		lastID:      0,
	}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	if sensor == nil {
		return errors.New("sensor is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if existingSensor, ok := r.sensorsBySN[sensor.SerialNumber]; ok && existingSensor.ID != sensor.ID {
		return errors.New("sensor with this serial number already exists")
	}
	if sensor.ID == 0 {
		r.lastID++
		sensor.ID = r.lastID

		if sensor.RegisteredAt.IsZero() {
			sensor.RegisteredAt = time.Now()
		}

	}

	r.sensors[sensor.ID] = sensor
	r.sensorsBySN[sensor.SerialNumber] = sensor

	return nil
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	sensors := make([]domain.Sensor, 0, len(r.sensors))
	for _, sensor := range r.sensors {
		sensors = append(sensors, *sensor)
	}

	return sensors, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	sensor, ok := r.sensors[id]
	if !ok {
		return nil, usecase.ErrSensorNotFound
	}

	return sensor, nil
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	sensor, ok := r.sensorsBySN[sn]
	if !ok {
		return nil, usecase.ErrSensorNotFound
	}

	return sensor, nil
}
