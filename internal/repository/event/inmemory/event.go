package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

type EventRepository struct {
	events map[int64][]*domain.Event
	mu     sync.RWMutex
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		events: make(map[int64][]*domain.Event),
	}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if event == nil {
		return errors.New("event is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[event.SensorID] = append(r.events[event.SensorID], event)
	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	events, exists := r.events[id]
	if !exists || len(events) == 0 {
		return nil, usecase.ErrEventNotFound
	}

	var latestEvent *domain.Event
	var latestTime time.Time
	for _, event := range events {
		if latestEvent == nil || event.Timestamp.After(latestTime) {
			latestEvent = event
			latestTime = event.Timestamp
		}
	}

	return latestEvent, nil
}

func (r *EventRepository) GetEventsHistoryBySensorID(ctx context.Context, id int64, startDate, endDate time.Time) ([]domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	events, exists := r.events[id]
	if !exists || len(events) == 0 {
		return []domain.Event{}, nil
	}

	var result []domain.Event
	for _, event := range events {
		if (event.Timestamp.Equal(startDate) || event.Timestamp.After(startDate)) &&
			(event.Timestamp.Equal(endDate) || event.Timestamp.Before(endDate)) {
			result = append(result, *event)
		}
	}

	return result, nil
}
