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

func Test_sensor_RegisterSensor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("fail, sensor not valid", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().SaveSensor(ctx, gomock.Any()).Times(0)

		s := NewSensor(sr)

		_, err := s.RegisterSensor(ctx, &domain.Sensor{
			SerialNumber: "1234567890",
			Type:         "some",
		})
		assert.ErrorIs(t, err, ErrWrongSensorType)

		_, err = s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "123", // wrong, should be 10 digits
		})
		assert.ErrorIs(t, err, ErrWrongSensorSerialNumber)

		_, err = s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "123456789011", // wrong, should be 10 digits
		})
		assert.ErrorIs(t, err, ErrWrongSensorSerialNumber)
	})

	t.Run("fail, repository return an error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		expectedError := errors.New("some error")
		sr.EXPECT().GetSensorBySerialNumber(ctx, gomock.Any()).Return(nil, expectedError)

		s := NewSensor(sr)

		_, err := s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "1234567890",
		})

		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("fail, repository return an error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		expectedError := errors.New("some error")
		sr.EXPECT().GetSensorBySerialNumber(ctx, gomock.Any()).Return(nil, ErrSensorNotFound)
		sr.EXPECT().SaveSensor(ctx, gomock.Any()).Times(1).Return(expectedError)

		a := NewSensor(sr)

		_, err := a.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "1234567890",
		})

		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("ok, register ok", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensor := &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "1234567890",
			Description:  "some desc",
		}

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().SaveSensor(ctx, gomock.Any()).Times(1).DoAndReturn(func(_ context.Context, ss *domain.Sensor) error {
			assert.Empty(t, ss.RegisteredAt)
			assert.Empty(t, ss.LastActivity)
			assert.Equal(t, sensor.Description, ss.Description)
			assert.Equal(t, sensor.Type, ss.Type)
			assert.Equal(t, sensor.SerialNumber, ss.SerialNumber)

			ss.RegisteredAt = time.Now()
			ss.ID = 1

			return nil
		})
		sr.EXPECT().GetSensorBySerialNumber(ctx, sensor.SerialNumber).Return(nil, ErrSensorNotFound)

		s := NewSensor(sr)

		sensor, err := s.RegisterSensor(ctx, sensor)
		assert.NoError(t, err)

		assert.NotEmpty(t, sensor.RegisteredAt)
		assert.Equal(t, int64(1), sensor.ID)
	})

	t.Run("ok, register idempotency", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensor := &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "1234567890",
			Description:  "some desc",
		}

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().SaveSensor(ctx, gomock.Any()).Times(1).DoAndReturn(func(_ context.Context, ss *domain.Sensor) error {
			assert.Empty(t, ss.RegisteredAt)
			assert.Empty(t, ss.LastActivity)
			assert.Equal(t, sensor.Description, ss.Description)
			assert.Equal(t, sensor.Type, ss.Type)
			assert.Equal(t, sensor.SerialNumber, ss.SerialNumber)

			ss.RegisteredAt = time.Now()
			ss.ID = 1

			return nil
		})
		sr.EXPECT().GetSensorBySerialNumber(ctx, sensor.SerialNumber).Return(nil, ErrSensorNotFound)

		s := NewSensor(sr)

		_, err := s.RegisterSensor(ctx, sensor)
		assert.NoError(t, err)

		assert.NotEmpty(t, sensor.RegisteredAt)
		assert.Equal(t, int64(1), sensor.ID)

		sr.EXPECT().GetSensorBySerialNumber(ctx, sensor.SerialNumber).Return(sensor, nil)

		sensor2, err := s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeContactClosure,
			SerialNumber: "1234567890",
			Description:  "some desc 2 ",
		})
		assert.NoError(t, err)

		assert.Equal(t, sensor.ID, sensor2.ID)
		assert.Equal(t, sensor.RegisteredAt, sensor2.RegisteredAt)
		assert.Equal(t, sensor.Description, sensor2.Description)
		assert.Equal(t, sensor.Type, sensor2.Type)
		assert.Equal(t, sensor.SerialNumber, sensor2.SerialNumber)
	})
}

func Test_sensor_GetSensors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("err, got err form repo", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		expectedError := errors.New("some error")
		sr.EXPECT().GetSensors(ctx).Times(1).Return(nil, expectedError)

		s := NewSensor(sr)

		_, err := s.GetSensors(ctx)
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("ok, got list", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensors(ctx).Times(1).Return([]domain.Sensor{
			{},
			{},
		}, nil)

		s := NewSensor(sr)

		list, err := s.GetSensors(ctx)
		assert.NoError(t, err)
		assert.Len(t, list, 2)
	})
}

func Test_sensor_GetSensorByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("err, got err form repo", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		expectedError := errors.New("some error")
		sr.EXPECT().GetSensorByID(ctx, gomock.Any()).Times(1).Return(nil, expectedError)

		s := NewSensor(sr)

		_, err := s.GetSensorByID(ctx, 1)
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("err, sensor not forund", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, gomock.Any()).Times(1).Return(nil, ErrSensorNotFound)

		s := NewSensor(sr)

		_, err := s.GetSensorByID(ctx, 1)
		assert.ErrorIs(t, err, ErrSensorNotFound)
	})

	t.Run("ok, got sensor", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, gomock.Any()).Times(1).Return(&domain.Sensor{
			ID:           1,
			SerialNumber: "0123456789",
			Type:         domain.SensorTypeADC,
			CurrentState: 255,
			Description:  "some desc",
			IsActive:     true,
			RegisteredAt: time.Now(),
		}, nil)

		s := NewSensor(sr)

		sensor, err := s.GetSensorByID(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, sensor)
	})
}
