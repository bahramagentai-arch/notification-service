package test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	notification "github.com/BahramRousta/notification-service/internal/notification/domain"
	pgstore "github.com/BahramRousta/notification-service/internal/notification/store"
)


var testDSN string

func setupStore(t *testing.T, dsn string) *pgstore.Store {
	t.Helper()
	ctx := context.Background()

	store, err := pgstore.NewStore(ctx, dsn)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(store.Close)
	return store
}


func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5454/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "notifications_test",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2). // Postgres logs this twice on startup
			WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	if err != nil {
		panic(fmt.Sprintf("start postgres container: %v", err))
	}
	defer container.Terminate(ctx) //nolint:errcheck
 
	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432/tcp")
 
	testDSN = fmt.Sprintf(
		"postgres://test:test@%s:%s/notifications_test?sslmode=disable",
		host, port.Port(),
	)
	_, thisFile, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(thisFile), "../../../../migrations")
 
	mg, err := migrate.New("file://"+migrationsPath, testDSN)
	if err != nil {
		panic(fmt.Sprintf("migrate.New: %v", err))
	}
	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(fmt.Sprintf("migrate up: %v", err))
	}
 
	m.Run()
}

func truncate(t *testing.T, store *pgstore.Store) {
	t.Helper()
	// We expose a test-only helper via the Store — see store_test_helper.go.
	if err := store.TruncateForTest(context.Background()); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}


func newMessage(overrides ...func(*notification.Message)) *notification.Message {
	msg := &notification.Message{
		ID:             uuid.NewString(),
		IdempotencyKey: uuid.NewString(),
		Recipient:      "+989121234567",
		Body:           "Hello from test",
		Channel:        notification.ChannelSMS,
		Priority:       notification.PriorityNormal,
		Status:         notification.StatusPending,
		Attempts:       0,
		MaxAttempts:    3,
		CreatedAt:      time.Now().UTC().Truncate(time.Millisecond),
		UpdatedAt:      time.Now().UTC().Truncate(time.Millisecond),
	}
	for _, fn := range overrides {
		fn(msg)
	}
	return msg
}