package pg_test

//nolint: revive // test stub
import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DbName = "db"
	DbUser = "postgres"
	DbPass = "postgres"
)

type TestDatabase struct {
	DbInstance *pgxpool.Pool
	DbAddress  string
	container  testcontainers.Container
}

func SetupTestDatabase() *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	container, dbInstance, dbAddr, err := createContainer(ctx)
	if err != nil {
		log.Fatal("failed to setup test", err)
	}

	err = migrateDb(dbAddr)
	if err != nil {
		log.Fatal("failed to perform db migration", err)
	}
	cancel()

	return &TestDatabase{
		container:  container,
		DbInstance: dbInstance,
		DbAddress:  dbAddr,
	}
}

func (tdb *TestDatabase) TearDown() {
	tdb.DbInstance.Close()
	_ = tdb.container.Terminate(context.Background())
}

func createContainer(ctx context.Context) (testcontainers.Container, *pgxpool.Pool, string, error) {
	env := map[string]string{
		"POSTGRES_PASSWORD": DbPass,
		"POSTGRES_USER":     DbUser,
		"POSTGRES_DB":       DbName,
	}

	port := "5432/tcp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16.2-alpine",
			ExposedPorts: []string{port},
			Env:          env,
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to start container: %w", err)
	}

	p, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to get container external port: %w", err)
	}

	log.Println("postgres container ready and running at port: ", p.Port())

	time.Sleep(time.Second)

	dbAddr := fmt.Sprintf("localhost:%s", p.Port())

	config, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", DbUser, DbPass, dbAddr, DbName))
	if err != nil {
		log.Fatalf("can't parse pgxpool config")
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return container, db, dbAddr, fmt.Errorf("failed to establish database connection: %w", err)
	}

	return container, db, dbAddr, nil
}

func migrateDb(dbAddr string) error {
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get path")
	}
	pathToMigrationFiles := filepath.Dir(path) + "/../../migrations"

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", DbUser, DbPass, dbAddr, DbName)
	m, err := migrate.New(fmt.Sprintf("file:%s", pathToMigrationFiles), databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Println("migration done")

	return nil
}
