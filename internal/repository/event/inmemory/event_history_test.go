package inmemory

import (
	"context"
	"homework/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventRepository_GetEventsHistoryBySensorID(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := er.GetEventsHistoryBySensorID(ctx, 0, time.Now().Add(-24*time.Hour), time.Now())
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := er.GetEventsHistoryBySensorID(ctx, 0, time.Now().Add(-24*time.Hour), time.Now())
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("empty result for non-existent sensor", func(t *testing.T) {
		er := NewEventRepository()
		ctx := context.Background()

		history, err := er.GetEventsHistoryBySensorID(ctx, 999, time.Now().Add(-24*time.Hour), time.Now())
		assert.NoError(t, err)
		assert.Empty(t, history)
	})

	t.Run("empty result for sensor with no events in time range", func(t *testing.T) {
		er := NewEventRepository()
		ctx := context.Background()
		sensorID := int64(123)

		event := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "1234567890",
			Payload:            42,
			Timestamp:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		assert.NoError(t, er.SaveEvent(ctx, event))

		startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

		history, err := er.GetEventsHistoryBySensorID(ctx, sensorID, startDate, endDate)
		assert.NoError(t, err)
		assert.Empty(t, history)
	})

	t.Run("returns events within time range", func(t *testing.T) {
		er := NewEventRepository()
		ctx := context.Background()
		sensorID := int64(123)

		eventBefore := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "1234567890",
			Payload:            10,
			Timestamp:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		eventInRange1 := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "1234567890",
			Payload:            20,
			Timestamp:          time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		}

		eventInRange2 := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "1234567890",
			Payload:            30,
			Timestamp:          time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		}

		eventAfter := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "1234567890",
			Payload:            40,
			Timestamp:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.NoError(t, er.SaveEvent(ctx, eventBefore))
		assert.NoError(t, er.SaveEvent(ctx, eventInRange1))
		assert.NoError(t, er.SaveEvent(ctx, eventInRange2))
		assert.NoError(t, er.SaveEvent(ctx, eventAfter))

		startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

		history, err := er.GetEventsHistoryBySensorID(ctx, sensorID, startDate, endDate)
		assert.NoError(t, err)
		assert.Len(t, history, 2)

		var payloads []int64
		for _, event := range history {
			payloads = append(payloads, event.Payload)
		}
		assert.Contains(t, payloads, eventInRange1.Payload)
		assert.Contains(t, payloads, eventInRange2.Payload)
		assert.NotContains(t, payloads, eventBefore.Payload)
		assert.NotContains(t, payloads, eventAfter.Payload)
	})

	t.Run("returns events at boundary dates", func(t *testing.T) {
		er := NewEventRepository()
		ctx := context.Background()
		sensorID := int64(456)

		startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

		eventAtStart := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "1234567890",
			Payload:            50,
			Timestamp:          startDate,
		}

		eventAtEnd := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "1234567890",
			Payload:            60,
			Timestamp:          endDate,
		}

		assert.NoError(t, er.SaveEvent(ctx, eventAtStart))
		assert.NoError(t, er.SaveEvent(ctx, eventAtEnd))

		history, err := er.GetEventsHistoryBySensorID(ctx, sensorID, startDate, endDate)
		assert.NoError(t, err)
		assert.Len(t, history, 2)

		var payloads []int64
		for _, event := range history {
			payloads = append(payloads, event.Payload)
		}
		assert.Contains(t, payloads, eventAtStart.Payload)
		assert.Contains(t, payloads, eventAtEnd.Payload)
	})
}
