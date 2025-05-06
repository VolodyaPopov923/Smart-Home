package usecase

import (
	"context"
	"homework/internal/domain"
)

type User struct {
	userRepo        UserRepository
	sensorOwnerRepo SensorOwnerRepository
	sensorRepo      SensorRepository
}

func NewUser(ur UserRepository, sor SensorOwnerRepository, sr SensorRepository) *User {
	return &User{
		userRepo:        ur,
		sensorOwnerRepo: sor,
		sensorRepo:      sr,
	}
}

func (u *User) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if user == nil {
		return nil, ErrUserNotFound
	}

	if user.Name == "" {
		return nil, ErrInvalidUserName
	}

	if err := u.userRepo.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *User) AttachSensorToUser(ctx context.Context, userID, sensorID int64) error {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	sensor, err := u.sensorRepo.GetSensorByID(ctx, sensorID)
	if err != nil {
		return err
	}
	if sensor == nil {
		return ErrSensorNotFound
	}
	sensorOwner := domain.SensorOwner{
		UserID:   userID,
		SensorID: sensorID,
	}

	return u.sensorOwnerRepo.SaveSensorOwner(ctx, sensorOwner)
}

func (u *User) GetUserSensors(ctx context.Context, userID int64) ([]domain.Sensor, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	sensorOwners, err := u.sensorOwnerRepo.GetSensorsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var sensors []domain.Sensor
	for _, sensorOwner := range sensorOwners {
		sensor, err := u.sensorRepo.GetSensorByID(ctx, sensorOwner.SensorID)
		if err != nil {
			return nil, err
		}
		if sensor != nil {
			sensors = append(sensors, *sensor)
		}
	}
	return sensors, nil
}
