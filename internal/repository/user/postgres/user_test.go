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

type UserTestSuite struct {
	suite.Suite
	testDbInstance *pgxpool.Pool
	testDB         *pg_test.TestDatabase

	repo *UserRepository
}

func (suite *UserTestSuite) SetupSuite() {
	suite.testDB = pg_test.SetupTestDatabase()
	suite.testDbInstance = suite.testDB.DbInstance

	suite.repo = NewUserRepository(suite.testDbInstance)
}

func (suite *UserTestSuite) TearDownSuite() {
	suite.testDB.TearDown()
}

func (suite *UserTestSuite) TestUserRepository_SaveUser() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	name := "vasya pupkin"

	err := suite.repo.SaveUser(ctx, &domain.User{
		Name: name,
	})

	assert.Nil(suite.T(), err)
}

func (suite *UserTestSuite) TestUserRepository_GetUserByID() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint: govet // test stub

	name := "vasya pupkin"

	err := suite.repo.SaveUser(ctx, &domain.User{
		Name: name,
	})

	assert.Nil(suite.T(), err)

	user, err := suite.repo.GetUserByID(ctx, 1)

	assert.Nil(suite.T(), err)

	assert.Equal(suite.T(), name, user.Name)
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
