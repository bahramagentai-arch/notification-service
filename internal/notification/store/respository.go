package store

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/BahramRousta/notification-service/internal/notification/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository interface {
	Create(ctx context.Context, msg *domain.Message) error
	
}

func (s *Store) Create(ctx context.Context, msg *domain.Message) error {
	const q = `
		INSERT INTO messages (
			id, idempotency_key, recipient, body,
			channel, priority, status,
			attempts, max_attempts,
			provider_msg_id, failure_reason,
			created_at, updated_at
		) VALUES (
			@id, @idempotency_key, @recipient, @body,
			@channel, @priority, @status,
			@attempts, @max_attempts,
			@provider_msg_id, @failure_reason,
			@created_at, @updated_at
		)`

	args := pgx.NamedArgs{
		"id":              msg.ID,
		"idempotency_key": msg.IdempotencyKey,
		"recipient":       msg.Recipient,
		"body":            msg.Body,
		"channel":         msg.Channel, "priority": msg.Priority,
		"status":          msg.Status,
		"attempts":        msg.Attempts,
		"max_attempts":    msg.MaxAttempts,
		"provider_msg_id": msg.ProviderMsgID,
		"failure_reason":  msg.FailureReason,
		"created_at":      msg.CreatedAt,
		"updated_at":      msg.UpdatedAt,
	}

	_, err := s.pool.Exec(ctx, q, args)
	if err != nil {
		if isDuplicateKey(err) {
			return fmt.Errorf("%w: key=%s", domain.ErrDuplicate, msg.IdempotencyKey)
		}
		return fmt.Errorf("postgres: create message: %w", err)
	}
	return nil
}


func isDuplicateKey(err error)bool{
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}