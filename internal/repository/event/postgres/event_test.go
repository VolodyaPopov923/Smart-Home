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

type EventTestSuite struct {
	suite.Suite
	testDbInstance *pgxpool.Pool
	testDB         *pg_test.TestDatabase

	repo *EventRepository
}

func (suite *EventTestSuite) SetupSuite() {
	suite.testDB = pg_test.SetupTestDatabase()
	suite.testDbInstance = suite.testDB.DbInstance

	suite.repo = NewEventRepository(suite.testDbInstance)
}

func (suite *EventTestSuite) TearDownSuite() {
	suite.testDB.TearDown()
}

func (suite *EventTestSuite) TestEventRepository_SaveEvent() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	err := suite.repo.SaveEvent(ctx, &domain.Event{
		Timestamp:          time.Now().In(time.UTC),
		SensorSerialNumber: "1234567890",
		SensorID:           1,
		Payload:            1,
	})

	assert.Nil(suite.T(), err)
}

func (suite *EventTestSuite) TestEventRepository_GetLastEventBySensorID() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	firstEvent := domain.Event{
		Timestamp:          time.Now().Truncate(time.Microsecond).In(time.UTC),
		SensorSerialNumber: "0987654321",
		SensorID:           2,
		Payload:            1,
	}

	secondEvent := domain.Event{
		Timestamp:          time.Now().Truncate(time.Microsecond).Add(time.Minute * 10).In(time.UTC),
		SensorSerialNumber: "0987654321",
		SensorID:           2,
		Payload:            2,
	}

	err := suite.repo.SaveEvent(ctx, &firstEvent)
	assert.Nil(suite.T(), err)

	err = suite.repo.SaveEvent(ctx, &secondEvent)
	assert.Nil(suite.T(), err)

	event, err := suite.repo.GetLastEventBySensorID(ctx, 2)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), secondEvent, *event)
}

func TestEventTestSuite(t *testing.T) {
	suite.Run(t, new(EventTestSuite))
}
