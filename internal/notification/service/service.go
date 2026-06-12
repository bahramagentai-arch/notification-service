package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/BahramRousta/notification-service/internal/notification/domain"
	"github.com/BahramRousta/notification-service/internal/notification/store"
	"github.com/google/uuid"
)

const defaultMaxAttempts = 3
const maxBodyRunes = 1600

type NotificationService struct {
	repo   store.Repository
	logger *slog.Logger
}

func New(repo store.Repository, logger *slog.Logger) *NotificationService {
	return &NotificationService{
		repo:   repo,
		logger: logger,
	}
}

type SendRequest struct {
	IdempotencyKey string
	Recipient      string
	Body           string
	Priority       domain.Priority
}

func (s *NotificationService) Send(ctx context.Context, req SendRequest) (*domain.Message, error) {

	now := time.Now().UTC()
	msg := &domain.Message{
		// uuid.NewString() generates a v4 UUID — random and globally unique.
		ID:             uuid.NewString(),
		IdempotencyKey: req.IdempotencyKey,
		Recipient:      req.Recipient,
		Body:           req.Body,
		Channel:        domain.ChannelSMS,
		Priority:       req.Priority,
		Status:         domain.StatusPending, // always starts as PENDING
		Attempts:       0,
		MaxAttempts:    defaultMaxAttempts,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("service: create message: %w", err)
	}
	s.logger.Info("message created, queued for delivery",
		slog.String("msg_id", msg.ID),
		slog.String("recipient", msg.Recipient),
		slog.Int("priority", int(msg.Priority)),
	)
	return msg, nil
}
