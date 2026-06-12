package test

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/BahramRousta/notification-service/internal/notification/domain"
	"github.com/BahramRousta/notification-service/internal/notification/providers/fake"
)

func TestFakeProvider_Send_ReturnNonEmptyID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	p := fake.New(logger)

	msg := &domain.Message{
		ID:        "test-id-01",
		Recipient: "+989121234455",
		Body:      "test body",
		Channel:   domain.ChannelSMS,
	}

	id, err := p.Send(context.Background(), msg)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if id == "" {
		t.Fatal("expected a non-empty provider message ID")
	}

	if !strings.HasPrefix(id, "fake-") {
		t.Errorf("expected ID to start with 'fake-', got: %s", id)
	}

}


func TestFakeProvider_Send_TwoCalls_DifferentIDs(t *testing.T){
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	p := fake.New(logger)
 
	msg := &domain.Message{Recipient: "+989121234567", Body: "hi"}
 
	id1, _ := p.Send(context.Background(), msg)
	id2, _ := p.Send(context.Background(), msg)
 
	if id1 == id2 {
		t.Errorf("expected different IDs for two sends, both got: %s", id1)
	}
}