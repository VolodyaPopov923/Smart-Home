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

func Test_event_GetLastEventBySensorID_Additional(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("err, repository error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		er := NewMockEventRepository(ctrl)
		expectedError := errors.New("database error")
		er.EXPECT().GetLastEventBySensorID(ctx, int64(1)).Return(nil, expectedError)

		e := NewEvent(er, nil)

		event, err := e.GetLastEventBySensorID(ctx, 1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedError)
		assert.Nil(t, event)
	})

	t.Run("err, event not found", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		er := NewMockEventRepository(ctrl)
		er.EXPECT().GetLastEventBySensorID(ctx, int64(1)).Return(nil, nil)

		e := NewEvent(er, nil)

		event, err := e.GetLastEventBySensorID(ctx, 1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEventNotFound)
		assert.Nil(t, event)
	})

	t.Run("ok, success", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		expectedEvent := &domain.Event{
			SensorID:           1,
			SensorSerialNumber: "1234567890",
			Payload:            42,
			Timestamp:          time.Now(),
		}

		er := NewMockEventRepository(ctrl)
		er.EXPECT().GetLastEventBySensorID(ctx, int64(1)).Return(expectedEvent, nil)

		e := NewEvent(er, nil)

		event, err := e.GetLastEventBySensorID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expectedEvent, event)
	})
}

func Test_event_GetSensorHistory_WithEdgeCases_Additional(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("err, get sensor by id returns nil with no error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, int64(1)).Return(nil, nil)

		e := NewEvent(nil, sr)

		events, err := e.GetSensorHistory(ctx, 1, time.Now(), time.Now())
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSensorNotFound)
		assert.Nil(t, events)
	})

	t.Run("success with empty response", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

		e := NewEvent(nil, sr)

		events, err := e.GetSensorHistory(ctx, 1, time.Now().Add(-24*time.Hour), time.Now())
		assert.NoError(t, err)
		assert.Empty(t, events)
	})
}

func Test_sensor_ValidationEdgeCases_Additional(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("validation of nil sensor", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		s := NewSensor(sr)

		_, err := s.RegisterSensor(ctx, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrWrongSensorSerialNumber)
	})

	t.Run("validation of various invalid serial numbers", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		s := NewSensor(sr)

		_, err := s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "",
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrWrongSensorSerialNumber)

		_, err = s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "12345678ab",
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrWrongSensorSerialNumber)

		_, err = s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "123456789!",
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrWrongSensorSerialNumber)
	})

	t.Run("validation of valid sensor with contact closure type", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorBySerialNumber(ctx, "1234567890").Return(nil, ErrSensorNotFound)
		sr.EXPECT().SaveSensor(ctx, gomock.Any()).Return(nil)

		s := NewSensor(sr)

		sensor, err := s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeContactClosure,
			SerialNumber: "1234567890",
		})
		assert.NoError(t, err)
		assert.NotNil(t, sensor)
		assert.Equal(t, domain.SensorTypeContactClosure, sensor.Type)
	})

	t.Run("error handling when GetSensorBySerialNumber returns non-ErrSensorNotFound error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		expectedError := errors.New("network error")
		sr.EXPECT().GetSensorBySerialNumber(ctx, "1234567890").Return(nil, expectedError)

		s := NewSensor(sr)

		_, err := s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "1234567890",
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedError)
	})
}

func Test_user_EdgeCases_Additional(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("register nil user", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		u := NewUser(nil, nil, nil)

		_, err := u.RegisterUser(ctx, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("get user sensors with empty result from GetSensorsByUserID", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, int64(1)).Return(&domain.User{ID: 1}, nil)

		sor := NewMockSensorOwnerRepository(ctrl)
		sor.EXPECT().GetSensorsByUserID(ctx, int64(1)).Return([]domain.SensorOwner{}, nil)

		u := NewUser(ur, sor, nil)

		sensors, err := u.GetUserSensors(ctx, 1)
		assert.NoError(t, err)
		assert.Empty(t, sensors)
	})

	t.Run("get user sensors with nil sensor returned by GetSensorByID", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, int64(1)).Return(&domain.User{ID: 1}, nil)

		sor := NewMockSensorOwnerRepository(ctrl)
		sor.EXPECT().GetSensorsByUserID(ctx, int64(1)).Return([]domain.SensorOwner{
			{UserID: 1, SensorID: 1},
			{UserID: 1, SensorID: 2},
		}, nil)

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, int64(1)).Return(nil, nil)
		sr.EXPECT().GetSensorByID(ctx, int64(2)).Return(&domain.Sensor{ID: 2}, nil)

		u := NewUser(ur, sor, sr)

		sensors, err := u.GetUserSensors(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, sensors, 1)
		assert.Equal(t, int64(2), sensors[0].ID)
	})
}

func Test_sensor_RegisterWithEmptyFields_Additional(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("testing sensor with valid ID but no other fields", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		s := NewSensor(sr)

		sr.EXPECT().GetSensorBySerialNumber(ctx, "1234567890").Return(nil, ErrSensorNotFound)
		sr.EXPECT().SaveSensor(ctx, gomock.Any()).Return(nil)

		_, err := s.RegisterSensor(ctx, &domain.Sensor{
			Type:         domain.SensorTypeADC,
			SerialNumber: "1234567890",
		})
		assert.NoError(t, err)
	})
}

func Test_user_GetUserSensorsWithMultipleSensors_Additional(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("get user sensors with mix of nil and valid sensors", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, int64(1)).Return(&domain.User{ID: 1}, nil)

		sor := NewMockSensorOwnerRepository(ctrl)
		sor.EXPECT().GetSensorsByUserID(ctx, int64(1)).Return([]domain.SensorOwner{
			{UserID: 1, SensorID: 1},
			{UserID: 1, SensorID: 2},
			{UserID: 1, SensorID: 3},
			{UserID: 1, SensorID: 4},
		}, nil)

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, int64(1)).Return(&domain.Sensor{ID: 1, Type: domain.SensorTypeADC}, nil)
		sr.EXPECT().GetSensorByID(ctx, int64(2)).Return(nil, nil)
		sr.EXPECT().GetSensorByID(ctx, int64(3)).Return(&domain.Sensor{ID: 3, Type: domain.SensorTypeContactClosure}, nil)
		sr.EXPECT().GetSensorByID(ctx, int64(4)).Return(nil, errors.New("not found"))

		u := NewUser(ur, sor, sr)

		_, err := u.GetUserSensors(ctx, 1)
		assert.Error(t, err)
		assert.Equal(t, "not found", err.Error())
	})
}

func Test_AllEdgeCases_Additional(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("event with zero sensorID", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		er := NewMockEventRepository(ctrl)
		er.EXPECT().GetLastEventBySensorID(ctx, int64(0)).Return(nil, nil)

		e := NewEvent(er, nil)

		_, err := e.GetLastEventBySensorID(ctx, 0)
		assert.Error(t, err)
		assert.Equal(t, ErrEventNotFound, err)
	})

	t.Run("sensor with edged types", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)
		s := NewSensor(sr)

		_, err := s.RegisterSensor(ctx, &domain.Sensor{
			SerialNumber: "1234567890",
			Type:         "",
		})
		assert.Error(t, err)
		assert.Equal(t, ErrWrongSensorType, err)

		_, err = s.RegisterSensor(ctx, &domain.Sensor{
			SerialNumber: "1234567890",
			Type:         "unknown",
		})
		assert.Error(t, err)
		assert.Equal(t, ErrWrongSensorType, err)
	})
}
