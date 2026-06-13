package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/BahramRousta/notification-service/internal/notification/domain"
	"github.com/BahramRousta/notification-service/internal/notification/store"
	"github.com/google/uuid"
)

const defaultMaxAttempts = 3
const maxBodyRunes = 1600


type Service interface {
	Send(ctx context.Context, req SendRequest) (*domain.Message, error)
}

type NotificationService struct {
	repo   store.Repository
	logger *slog.Logger
}

func NewService(repo store.Repository, logger *slog.Logger) *NotificationService {
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

	if err := validateRequest(req); err != nil {
		return nil, err
	}

	if req.IdempotencyKey != ""{
		existing, err := s.repo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil {
			s.logger.Info("idempotency key already exists, returning existing message",
				slog.String("idempotency_key", req.IdempotencyKey),
				slog.String("msg_id", existing.ID),
				slog.String("status", string(existing.Status)),
			)
			return existing, nil
		}

		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("service: check idempotency key: %w", err)
		}
	}
	now := time.Now().UTC()
	msg := &domain.Message{
		ID:             uuid.NewString(),
		IdempotencyKey: req.IdempotencyKey,
		Recipient:      req.Recipient,
		Body:           req.Body,
		Channel:        domain.ChannelSMS,
		Priority:       req.Priority,
		Status:         domain.StatusPending,
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
