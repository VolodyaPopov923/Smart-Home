package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_event_ReceiveEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("err, invalid event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		e := NewEvent(nil, nil)

		err := e.ReceiveEvent(ctx, &domain.Event{})
		assert.ErrorIs(t, err, ErrInvalidEventTimestamp)
	})

	t.Run("err, sensor not found", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)

		sr.EXPECT().GetSensorBySerialNumber(ctx, gomock.Any()).Times(1).Return(nil, ErrSensorNotFound)

		e := NewEvent(nil, sr)

		err := e.ReceiveEvent(ctx, &domain.Event{
			Timestamp: time.Now(),
		})
		assert.ErrorIs(t, err, ErrSensorNotFound)
	})

	t.Run("err, event save error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)

		sr.EXPECT().GetSensorBySerialNumber(ctx, "0123456789").Times(1).Return(&domain.Sensor{
			ID: 1,
		}, nil)

		er := NewMockEventRepository(ctrl)
		expectedError := errors.New("some error")
		er.EXPECT().SaveEvent(ctx, gomock.Any()).Times(1).Return(expectedError)

		e := NewEvent(er, sr)

		err := e.ReceiveEvent(ctx, &domain.Event{
			Timestamp:          time.Now(),
			SensorSerialNumber: "0123456789",
		})
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("err, sensor save error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)

		sr.EXPECT().GetSensorBySerialNumber(ctx, "0123456789").Times(1).Return(&domain.Sensor{
			ID: 1,
		}, nil)
		expectedError := errors.New("some error")
		sr.EXPECT().SaveSensor(ctx, gomock.Any()).Times(1).Times(1).Return(expectedError)

		er := NewMockEventRepository(ctrl)
		er.EXPECT().SaveEvent(ctx, gomock.Any()).Times(1).Return(nil)

		e := NewEvent(er, sr)

		err := e.ReceiveEvent(ctx, &domain.Event{
			Timestamp:          time.Now(),
			SensorSerialNumber: "0123456789",
		})
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("ok, no error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)

		sr.EXPECT().GetSensorBySerialNumber(ctx, "0123456789").Times(1).Return(&domain.Sensor{
			ID: 1,
		}, nil)
		sr.EXPECT().SaveSensor(ctx, gomock.Any()).Times(1).Do(func(_ context.Context, s *domain.Sensor) {
			assert.Equal(t, int64(8), s.CurrentState)
			assert.NotEmpty(t, s.LastActivity)
		})

		er := NewMockEventRepository(ctrl)
		er.EXPECT().SaveEvent(ctx, gomock.Any()).Times(1).DoAndReturn(func(_ context.Context, event *domain.Event) error {
			assert.Equal(t, int64(1), event.SensorID)
			assert.Equal(t, "0123456789", event.SensorSerialNumber)

			return nil
		})

		e := NewEvent(er, sr)
		err := e.ReceiveEvent(ctx, &domain.Event{
			Timestamp:          time.Now(),
			SensorSerialNumber: "0123456789",
			Payload:            8,
		})
		assert.NoError(t, err)
	})
}
