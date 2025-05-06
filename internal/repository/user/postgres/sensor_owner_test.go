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

type SensorOwnerTestSuite struct {
	suite.Suite
	testDbInstance *pgxpool.Pool
	testDB         *pg_test.TestDatabase

	repo *SensorOwnerRepository
}

func (suite *SensorOwnerTestSuite) SetupSuite() {
	suite.testDB = pg_test.SetupTestDatabase()
	suite.testDbInstance = suite.testDB.DbInstance

	suite.repo = NewSensorOwnerRepository(suite.testDbInstance)
}

func (suite *SensorOwnerTestSuite) TearDownSuite() {
	suite.testDB.TearDown()
}

func (suite *SensorOwnerTestSuite) TestSensorOwnerRepository_SaveSensorOwner() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	err := suite.repo.SaveSensorOwner(ctx, domain.SensorOwner{
		UserID:   1,
		SensorID: 1,
	})

	assert.Nil(suite.T(), err)
}

func (suite *SensorOwnerTestSuite) TestSensorOwnerRepository_GetSensorsByUserID() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	err := suite.repo.SaveSensorOwner(ctx, domain.SensorOwner{
		UserID:   2,
		SensorID: 2,
	})

	assert.Nil(suite.T(), err)

	err = suite.repo.SaveSensorOwner(ctx, domain.SensorOwner{
		UserID:   2,
		SensorID: 3,
	})

	assert.Nil(suite.T(), err)

	sensors, err := suite.repo.GetSensorsByUserID(ctx, 2)

	assert.Nil(suite.T(), err)

	assert.ElementsMatch(suite.T(), []domain.SensorOwner{
		{UserID: 2, SensorID: 2},
		{UserID: 2, SensorID: 3},
	}, sensors)
}

func TestSensorOwnerTestSuite(t *testing.T) {
	suite.Run(t, new(SensorOwnerTestSuite))
}
