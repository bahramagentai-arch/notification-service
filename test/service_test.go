package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/BahramRousta/notification-service/internal/notification/domain"
	"github.com/BahramRousta/notification-service/internal/notification/service"
	"github.com/google/uuid"
)


func validRequest () service.SendRequest{
	return service.SendRequest{
		IdempotencyKey: uuid.NewString(),
		Recipient: "+989121112233",
		Body: "test",
		Priority: domain.PriorityNormal,
	}
}

func TestService_CreateMsgWithPendingStatus(t *testing.T) {
	store:= setupStore(t)
	defer truncate(t, store)
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	svc := service.NewService(store, logger)

	request := validRequest()
	
	msg, err := svc.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send: unexpected error: %v", err)
	}

	if msg.ID == "" {
		t.Error("Send: expected a non-empty message ID")
	}

	if msg.IdempotencyKey != request.IdempotencyKey {
		t.Fatalf("both key must be same")
	}

	if msg.Status != domain.StatusPending {
		t.Errorf("Send: status want PENDING, got %s:", msg.Status)
	}
	
}