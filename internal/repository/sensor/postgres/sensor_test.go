package postgres

import (
	"context"
	"homework/internal/domain"
	"homework/pkg/pg_test"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SensorTestSuite struct {
	suite.Suite
	testDbInstance *pgxpool.Pool
	testDB         *pg_test.TestDatabase

	repo *SensorRepository
}

func (suite *SensorTestSuite) SetupSuite() {
	suite.testDB = pg_test.SetupTestDatabase()
	suite.testDbInstance = suite.testDB.DbInstance

	suite.repo = NewSensorRepository(suite.testDbInstance)
}

func (suite *SensorTestSuite) TearDownSuite() {
	suite.testDB.TearDown()
}

func (suite *SensorTestSuite) TestSensorRepository_SaveSensor() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	sn := "1234567890"

	now := time.Now().Truncate(time.Microsecond)

	time.Sleep(time.Millisecond * 10)
	err := suite.repo.SaveSensor(ctx, &domain.Sensor{
		SerialNumber: sn,
		Type:         domain.SensorTypeADC,
		CurrentState: 1,
		Description:  "test_desc",
		IsActive:     true,
		RegisteredAt: now,
		LastActivity: now,
	})

	assert.Nil(suite.T(), err)

	sensor, err := suite.repo.GetSensorBySerialNumber(ctx, sn)

	assert.Nil(suite.T(), err)
	assert.NotEqual(suite.T(), sensor.RegisteredAt, sensor.LastActivity)

	updatedSensor := domain.Sensor{
		ID:           sensor.ID,
		SerialNumber: sn,
		Type:         domain.SensorTypeADC,
		CurrentState: 2,
		Description:  "test_desc_2",
		IsActive:     false,
		RegisteredAt: time.Now().Truncate(time.Microsecond).In(time.UTC),
		LastActivity: time.Now().Truncate(time.Microsecond).In(time.UTC),
	}

	// update old sensor
	err = suite.repo.SaveSensor(ctx, &updatedSensor)

	assert.Nil(suite.T(), err)

	sensor, err = suite.repo.GetSensorBySerialNumber(ctx, sn)

	updatedSensor.RegisteredAt = sensor.RegisteredAt

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), updatedSensor, *sensor)
}

func (suite *SensorTestSuite) TestSensorRepository_GetSensors() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	sn := "0987654321"

	newSensor := domain.Sensor{
		SerialNumber: sn,
		Type:         domain.SensorTypeADC,
		CurrentState: 1,
		Description:  "test_desc_3",
		IsActive:     true,
		RegisteredAt: time.Now().Truncate(time.Microsecond).In(time.UTC),
		LastActivity: time.Now().Truncate(time.Microsecond).In(time.UTC),
	}
	err := suite.repo.SaveSensor(ctx, &newSensor)

	assert.Nil(suite.T(), err)

	sensor, err := suite.repo.GetSensorBySerialNumber(ctx, sn)

	assert.Nil(suite.T(), err)

	newSensor.ID = sensor.ID
	newSensor.RegisteredAt = sensor.RegisteredAt

	sensors, err := suite.repo.GetSensors(ctx)

	assert.Nil(suite.T(), err)
	assert.Contains(suite.T(), sensors, newSensor)
}

func (suite *SensorTestSuite) TestSensorRepository_GetSensorByID() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	sn := "1987654321"

	newSensor := domain.Sensor{
		SerialNumber: sn,
		Type:         domain.SensorTypeADC,
		CurrentState: 1,
		Description:  "test_desc_4",
		IsActive:     true,
		RegisteredAt: time.Now().Truncate(time.Microsecond).In(time.UTC),
		LastActivity: time.Now().Truncate(time.Microsecond).In(time.UTC),
	}
	err := suite.repo.SaveSensor(ctx, &newSensor)

	assert.Nil(suite.T(), err)

	sensor, err := suite.repo.GetSensorBySerialNumber(ctx, sn)

	assert.Nil(suite.T(), err)

	newSensor.ID = sensor.ID
	newSensor.RegisteredAt = sensor.RegisteredAt

	sensor, err = suite.repo.GetSensorByID(ctx, sensor.ID)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), newSensor, *sensor)
}

func (suite *SensorTestSuite) TestSensorRepository_GetSensorBySerialNumber() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	sn := "2987654321"

	newSensor := domain.Sensor{
		SerialNumber: sn,
		Type:         domain.SensorTypeADC,
		CurrentState: 1,
		Description:  "test_desc_5",
		IsActive:     true,
		RegisteredAt: time.Now().Truncate(time.Microsecond).In(time.UTC),
		LastActivity: time.Now().Truncate(time.Microsecond).In(time.UTC),
	}
	err := suite.repo.SaveSensor(ctx, &newSensor)

	assert.Nil(suite.T(), err)

	sensor, err := suite.repo.GetSensorBySerialNumber(ctx, sn)

	newSensor.ID = sensor.ID
	newSensor.RegisteredAt = sensor.RegisteredAt

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), newSensor, *sensor)
}

func TestSensorTestSuite(t *testing.T) {
	suite.Run(t, new(SensorTestSuite))
}
