package test

import (
	"context"
	"errors"
	"testing"

	notification "github.com/BahramRousta/notification-service/internal/notification/domain"
	"github.com/google/uuid"
)



func TestStore_Create(t *testing.T) {
	store := setupStore(t, testDSN)
	truncate(t, store)
	ctx := context.Background()
 
	t.Run("creates a message successfully", func(t *testing.T) {
		msg := newMessage()
		if err := store.Create(ctx, msg); err != nil {
			t.Fatalf("Create: %v", err)
		}
	})
 
	t.Run("returns ErrDuplicate on repeated idempotency key", func(t *testing.T) {
		key := uuid.NewString()
		msg1 := newMessage(func(m *notification.Message) { m.IdempotencyKey = key })
		msg2 := newMessage(func(m *notification.Message) { m.IdempotencyKey = key })
 
		if err := store.Create(ctx, msg1); err != nil {
			t.Fatalf("first Create: %v", err)
		}
		err := store.Create(ctx, msg2)
		if !errors.Is(err, notification.ErrDuplicate) {
			t.Fatalf("expected ErrDuplicate, got: %v", err)
		}
	})
}