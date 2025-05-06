package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
	"regexp"
)

type Sensor struct {
	sensorRepo SensorRepository
}

func NewSensor(sr SensorRepository) *Sensor {
	return &Sensor{
		sensorRepo: sr,
	}
}

func isSensorValid(sensor *domain.Sensor) error {
	if sensor == nil {
		return ErrWrongSensorSerialNumber
	}

	matched, _ := regexp.MatchString(`^\d{10}$`, sensor.SerialNumber)
	if !matched {
		return ErrWrongSensorSerialNumber
	}

	if sensor.Type != domain.SensorTypeADC && sensor.Type != domain.SensorTypeContactClosure {
		return ErrWrongSensorType
	}

	return nil
}

func (s *Sensor) RegisterSensor(ctx context.Context, sensor *domain.Sensor) (*domain.Sensor, error) {
	if err := isSensorValid(sensor); err != nil {
		return nil, err
	}

	existingSensor, err := s.sensorRepo.GetSensorBySerialNumber(ctx, sensor.SerialNumber)

	if err != nil && !errors.Is(err, ErrSensorNotFound) {
		return nil, err
	}

	if existingSensor != nil {
		return existingSensor, nil
	}

	if err := s.sensorRepo.SaveSensor(ctx, sensor); err != nil {
		return nil, err
	}

	return sensor, nil
}

func (s *Sensor) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	sensors, err := s.sensorRepo.GetSensors(ctx)
	if err != nil {
		return nil, err
	}

	return sensors, nil
}

func (s *Sensor) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	sensor, err := s.sensorRepo.GetSensorByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sensor == nil {
		return nil, ErrSensorNotFound
	}

	return sensor, nil
}
