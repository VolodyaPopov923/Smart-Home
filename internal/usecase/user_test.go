package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_user_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("fail, user not valid", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		u := NewUser(nil, nil, nil)

		_, err := u.RegisterUser(ctx, &domain.User{})
		assert.ErrorIs(t, err, ErrInvalidUserName)
	})

	t.Run("fail, repo fail", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		expectedError := errors.New("doh")
		ur.EXPECT().SaveUser(ctx, gomock.Any()).Times(1).Return(expectedError)

		u := NewUser(ur, nil, nil)

		_, err := u.RegisterUser(ctx, &domain.User{
			Name: "Homer Simpson",
		})
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("ok", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().SaveUser(ctx, gomock.Any()).Times(1).Do(func(_ context.Context, u *domain.User) {
			assert.Equal(t, "Homer Simpson", u.Name)
			u.ID = 1
		})

		u := NewUser(ur, nil, nil)

		user, err := u.RegisterUser(ctx, &domain.User{
			Name: "Homer Simpson",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
	})
}

func Test_user_AttachSensorToUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("fail, user not found", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, gomock.Any()).Times(1).Return(nil, ErrUserNotFound)

		u := NewUser(ur, nil, nil)

		err := u.AttachSensorToUser(ctx, 1, 1)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("fail, user not found", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, gomock.Any()).Times(1).Return(&domain.User{ID: 1}, nil)

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, gomock.Any()).Times(1).Return(nil, ErrSensorNotFound)

		u := NewUser(ur, nil, sr)

		err := u.AttachSensorToUser(ctx, 1, 1)
		assert.ErrorIs(t, err, ErrSensorNotFound)
	})

	t.Run("fail, save error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, gomock.Any()).Times(1).Return(&domain.User{ID: 1}, nil)

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, gomock.Any()).Times(1).Return(&domain.Sensor{ID: 1}, nil)

		sor := NewMockSensorOwnerRepository(ctrl)
		expectedError := errors.New("some error")
		sor.EXPECT().SaveSensorOwner(ctx, gomock.Any()).Times(1).Return(expectedError)

		u := NewUser(ur, sor, sr)

		err := u.AttachSensorToUser(ctx, 1, 1)
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("ok, save success", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, gomock.Any()).Times(1).Return(&domain.User{ID: 1}, nil)

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, gomock.Any()).Times(1).Return(&domain.Sensor{ID: 1}, nil)

		sor := NewMockSensorOwnerRepository(ctrl)
		sor.EXPECT().SaveSensorOwner(ctx, gomock.Any()).Times(1).Return(nil).Do(func(_ context.Context, o domain.SensorOwner) {
			assert.Equal(t, int64(1), o.UserID)
			assert.Equal(t, int64(1), o.SensorID)
		})

		u := NewUser(ur, sor, sr)

		err := u.AttachSensorToUser(ctx, 1, 1)
		assert.NoError(t, err)
	})
}

func Test_user_GetUserSensors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("fail, user not found", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, gomock.Any()).Times(1).Return(nil, ErrUserNotFound)

		u := NewUser(ur, nil, nil)

		_, err := u.GetUserSensors(ctx, 1)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("fail, sensors owner repo error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, gomock.Any()).Times(1).Return(&domain.User{ID: 1}, nil)

		sor := NewMockSensorOwnerRepository(ctrl)
		expectedError := errors.New("some error")
		sor.EXPECT().GetSensorsByUserID(ctx, gomock.Any()).Times(1).Return(nil, expectedError)

		u := NewUser(ur, sor, nil)

		_, err := u.GetUserSensors(ctx, 1)
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("fail, sensor repo error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, gomock.Any()).Times(1).Return(&domain.User{ID: 1}, nil)

		sor := NewMockSensorOwnerRepository(ctrl)
		sor.EXPECT().GetSensorsByUserID(ctx, gomock.Any()).Times(1).Return([]domain.SensorOwner{
			{
				UserID:   1,
				SensorID: 1,
			},
		}, nil)

		sr := NewMockSensorRepository(ctrl)
		expectedError := errors.New("some error")
		sr.EXPECT().GetSensorByID(ctx, gomock.Any()).Times(1).Return(nil, expectedError)

		u := NewUser(ur, sor, sr)

		_, err := u.GetUserSensors(ctx, 1)
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("fail, sensors owner repo error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ur := NewMockUserRepository(ctrl)
		ur.EXPECT().GetUserByID(ctx, gomock.Any()).Times(1).Return(&domain.User{ID: 1}, nil)

		sor := NewMockSensorOwnerRepository(ctrl)
		sor.EXPECT().GetSensorsByUserID(ctx, gomock.Any()).Times(1).Return([]domain.SensorOwner{
			{
				UserID:   1,
				SensorID: 1,
			},
			{
				UserID:   1,
				SensorID: 2,
			},
			{
				UserID:   1,
				SensorID: 3,
			},
		}, nil)

		sr := NewMockSensorRepository(ctrl)
		sr.EXPECT().GetSensorByID(ctx, int64(1)).Times(1).Return(&domain.Sensor{ID: 1, Type: domain.SensorTypeADC}, nil)
		sr.EXPECT().GetSensorByID(ctx, int64(2)).Times(1).Return(&domain.Sensor{ID: 2, Type: domain.SensorTypeContactClosure}, nil)
		sr.EXPECT().GetSensorByID(ctx, int64(3)).Times(1).Return(&domain.Sensor{ID: 3, Type: domain.SensorTypeContactClosure}, nil)

		u := NewUser(ur, sor, sr)

		sensors, err := u.GetUserSensors(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, sensors, 3)
	})
}
