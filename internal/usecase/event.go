package usecase

import (
	"context"
	"homework/internal/domain"
	"time"
)

type Event struct {
	eventRepo  EventRepository
	sensorRepo SensorRepository
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{
		eventRepo:  er,
		sensorRepo: sr,
	}
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	if event.Timestamp.IsZero() {
		return ErrInvalidEventTimestamp
	}

	sensor, err := e.sensorRepo.GetSensorBySerialNumber(ctx, event.SensorSerialNumber)
	if err != nil {
		return err
	}
	if sensor == nil {
		return ErrSensorNotFound
	}

	event.SensorID = sensor.ID

	if err := e.eventRepo.SaveEvent(ctx, event); err != nil {
		return err
	}
	sensor.CurrentState = event.Payload
	sensor.LastActivity = time.Now()

	return e.sensorRepo.SaveSensor(ctx, sensor)
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	event, err := e.eventRepo.GetLastEventBySensorID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}
	return event, nil
}

func (e *Event) GetSensorHistory(ctx context.Context, id int64, startDate, endDate time.Time) ([]domain.Event, error) {
	sensor, err := e.sensorRepo.GetSensorByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sensor == nil {
		return nil, ErrSensorNotFound
	}

	if historyRepo, ok := e.eventRepo.(interface {
		GetEventsHistoryBySensorID(ctx context.Context, id int64, startDate, endDate time.Time) ([]domain.Event, error)
	}); ok {
		return historyRepo.GetEventsHistoryBySensorID(ctx, id, startDate, endDate)
	}

	return []domain.Event{}, nil
}
