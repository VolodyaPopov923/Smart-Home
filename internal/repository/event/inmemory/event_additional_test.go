package inmemory

import (
	"context"
	"homework/internal/domain"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventRepository_SaveEvent_Additional(t *testing.T) {
	t.Run("save multiple events for different sensors", func(t *testing.T) {
		er := NewEventRepository()
		ctx := context.Background()

		sensorID1 := int64(111)
		sensorID2 := int64(222)

		event1 := &domain.Event{
			SensorID:           sensorID1,
			SensorSerialNumber: "1111111111",
			Payload:            11,
			Timestamp:          time.Now(),
		}

		event2 := &domain.Event{
			SensorID:           sensorID2,
			SensorSerialNumber: "2222222222",
			Payload:            22,
			Timestamp:          time.Now(),
		}

		assert.NoError(t, er.SaveEvent(ctx, event1))
		assert.NoError(t, er.SaveEvent(ctx, event2))

		eventSensor1, err := er.GetLastEventBySensorID(ctx, sensorID1)
		assert.NoError(t, err)
		assert.Equal(t, event1.Payload, eventSensor1.Payload)

		eventSensor2, err := er.GetLastEventBySensorID(ctx, sensorID2)
		assert.NoError(t, err)
		assert.Equal(t, event2.Payload, eventSensor2.Payload)
	})

	t.Run("concurrent save with many goroutines", func(t *testing.T) {
		er := NewEventRepository()
		ctx := context.Background()
		sensorID := int64(333)

		numGoroutines := 100
		wg := sync.WaitGroup{}
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(i int) {
				defer wg.Done()
				event := &domain.Event{
					SensorID:           sensorID,
					SensorSerialNumber: "3333333333",
					Payload:            int64(i),
					Timestamp:          time.Now(),
				}
				assert.NoError(t, er.SaveEvent(ctx, event))
			}(i)
		}

		wg.Wait()

		er.mu.RLock()
		events := er.events[sensorID]
		er.mu.RUnlock()
		assert.Len(t, events, numGoroutines)
	})
}

func TestEventRepository_GetLastEventBySensorID_Additional(t *testing.T) {
	t.Run("get latest event with exact same timestamps", func(t *testing.T) {
		er := NewEventRepository()
		ctx := context.Background()
		sensorID := int64(444)

		timestamp := time.Date(2025, 4, 21, 14, 2, 27, 0, time.UTC)

		event1 := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "4444444444",
			Payload:            101,
			Timestamp:          timestamp,
		}

		event2 := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "4444444444",
			Payload:            102,
			Timestamp:          timestamp,
		}

		assert.NoError(t, er.SaveEvent(ctx, event1))
		assert.NoError(t, er.SaveEvent(ctx, event2))

		lastEvent, err := er.GetLastEventBySensorID(ctx, sensorID)
		assert.NoError(t, err)

		assert.Equal(t, timestamp, lastEvent.Timestamp)
		assert.Contains(t, []int64{101, 102}, lastEvent.Payload)
	})

	t.Run("get latest event from multiple events", func(t *testing.T) {
		er := NewEventRepository()
		ctx := context.Background()
		sensorID := int64(555)

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				SensorID:           sensorID,
				SensorSerialNumber: "5555555555",
				Payload:            int64(i),
				Timestamp:          time.Date(2025, 4, 21, 14, 0, i, 0, time.UTC),
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		latestEvent := &domain.Event{
			SensorID:           sensorID,
			SensorSerialNumber: "5555555555",
			Payload:            999,
			Timestamp:          time.Date(2025, 4, 21, 15, 0, 0, 0, time.UTC),
		}
		assert.NoError(t, er.SaveEvent(ctx, latestEvent))

		lastEvent, err := er.GetLastEventBySensorID(ctx, sensorID)
		assert.NoError(t, err)

		assert.Equal(t, latestEvent.Timestamp, lastEvent.Timestamp)
		assert.Equal(t, latestEvent.Payload, lastEvent.Payload)
	})

	t.Run("get latest event with events in random order", func(t *testing.T) {
		er := NewEventRepository()
		ctx := context.Background()
		sensorID := int64(666)

		timestamps := []time.Time{
			time.Date(2025, 2, 15, 10, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 10, 8, 0, 0, 0, time.UTC),
			time.Date(2025, 4, 21, 14, 2, 27, 0, time.UTC),
			time.Date(2025, 3, 5, 12, 0, 0, 0, time.UTC),
		}

		latestTimestamp := timestamps[2]

		for i, ts := range timestamps {
			event := &domain.Event{
				SensorID:           sensorID,
				SensorSerialNumber: "6666666666",
				Payload:            int64(i + 1),
				Timestamp:          ts,
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		lastEvent, err := er.GetLastEventBySensorID(ctx, sensorID)
		assert.NoError(t, err)

		assert.Equal(t, latestTimestamp, lastEvent.Timestamp)
		assert.Equal(t, int64(3), lastEvent.Payload)
	})
}
