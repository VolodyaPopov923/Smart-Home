package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
)

var (
	ErrWrongSensorSerialNumber = errors.New("wrong sensor serial number")
	ErrWrongSensorType         = errors.New("wrong sensor type")
	ErrInvalidEventTimestamp   = errors.New("invalid event timestamp")
	ErrInvalidUserName         = errors.New("invalid user name")
	ErrSensorNotFound          = errors.New("sensor not found")
	ErrUserNotFound            = errors.New("user not found")
	ErrEventNotFound           = errors.New("event not found")
)

//go:generate mockgen -source usecase.go -package usecase -destination usecase_mock.go
type SensorRepository interface {
	// SaveSensor - функция сохранения датчика
	SaveSensor(ctx context.Context, sensor *domain.Sensor) error
	// GetSensors - функция получения списка датчиков
	GetSensors(ctx context.Context) ([]domain.Sensor, error)
	// GetSensorByID - функция получения датчика по ID
	GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error)
	// GetSensorBySerialNumber - функция получения датчика по серийному номеру
	GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error)
}

type EventRepository interface {
	// SaveEvent - функция сохранения события по датчику
	SaveEvent(ctx context.Context, event *domain.Event) error
	// GetLastEventBySensorID - функция получения последнего события по ID датчика
	GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error)
}

type UserRepository interface {
	// SaveUser - функция сохранения пользователя
	SaveUser(ctx context.Context, user *domain.User) error
	// GetUserByID - функция получения пользователя по id
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
}

type SensorOwnerRepository interface {
	// SaveSensorOwner - функция привязки датчика к пользователю
	SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error
	// GetSensorsByUserID -функция, возвращающая список привязок для пользователя
	GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error)
}
