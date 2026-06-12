package test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/BahramRousta/notification-service/internal/notification/domain"
	pgstore "github.com/BahramRousta/notification-service/internal/notification/store"
)

var testDSN string

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "notifications_test",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
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
	// Note: os.Exit skips defers, so we terminate the container explicitly
	// before calling os.Exit at the bottom.
	defer container.Terminate(ctx) //nolint:errcheck

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432/tcp")

	testDSN = fmt.Sprintf(
		"postgres://test:test@%s:%s/notifications_test?sslmode=disable",
		host, port.Port(),
	)
	log.Printf("[test] DSN: %s", testDSN)

	moduleRoot, err := findModuleRoot()
	if err != nil {
		panic(fmt.Sprintf("find module root: %v", err))
	}
	migrationsPath := filepath.Join(moduleRoot, "migrations")
	log.Printf("[test] migrations path: %s", migrationsPath)

	// Hard-fail if the directory doesn't exist — catches path resolution bugs
	// immediately instead of silently running against an empty schema.
	if _, err := os.Stat(migrationsPath); err != nil {
		panic(fmt.Sprintf("migrations directory not found at %s: %v", migrationsPath, err))
	}

	mg, err := migrate.New("file://"+migrationsPath, testDSN)
	if err != nil {
		panic(fmt.Sprintf("migrate.New: %v", err))
	}
	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(fmt.Sprintf("migrate up: %v", err))
	}
	log.Printf("[test] migrations applied successfully")

	os.Exit(m.Run())
}

// findModuleRoot walks up from cwd until it finds the directory containing
// go.mod. Works regardless of where GoLand places the compiled test binary.
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	log.Printf("[test] searching for go.mod from cwd: %s", dir)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			log.Printf("[test] found go.mod at: %s", dir)
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s upward", dir)
		}
		dir = parent
	}
}

// ── Shared helpers ────────────────────────────────────────────────────────────

func setupStore(t *testing.T) *pgstore.Store {
	t.Helper()
	store, err := pgstore.NewStore(context.Background(), testDSN)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(store.Close)
	return store
}

func truncate(t *testing.T, store *pgstore.Store) {
	t.Helper()
	if err := store.TruncateForTest(context.Background()); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func newMessage(overrides ...func(*domain.Message)) *domain.Message {
	msg := &domain.Message{
		ID:             uuid.NewString(),
		IdempotencyKey: uuid.NewString(),
		Recipient:      "+989121234567",
		Body:           "Hello from test",
		Channel:        domain.ChannelSMS,
		Priority:       domain.PriorityNormal,
		Status:         domain.StatusPending,
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
