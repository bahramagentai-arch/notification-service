package store

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/BahramRousta/notification-service/internal/notification/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const msgColumns = `
	id, idempotency_key, recipient, body,
	channel, priority, status,
	attempts, max_attempts,
	provider_msg_id, failure_reason,
	created_at, updated_at`

type Repository interface {
	Create(ctx context.Context, msg *domain.Message) error
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Message, error)
	GetByID(ctx context.Context, id string) (*domain.Message, error)
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

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

type scanner interface {
	Scan(dest ...any) error
}

func scanMessage(s scanner) (*domain.Message, error) {
	var (
		msg     domain.Message
		channel string
		status  string
		prio    int16
		att     int16
		maxAtt  int16
	)
	err := s.Scan(
		&msg.ID, &msg.IdempotencyKey, &msg.Recipient, &msg.Body,
		&channel, &prio, &status,
		&att, &maxAtt,
		&msg.ProviderMsgID, &msg.FailureReason,
		&msg.CreatedAt, &msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	msg.Channel = domain.Channel(channel)
	msg.Status = domain.Status(status)
	msg.Priority = domain.Priority(prio)
	msg.Attempts = int(att)
	msg.MaxAttempts = int(maxAtt)
	return &msg, nil
}

func (s *Store) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Message, error) {
	const q = `SELECT ` + msgColumns + ` FROM messages WHERE idempotency_key = @key ;`
	row := s.pool.QueryRow(ctx, q, pgx.NamedArgs{"key": key})
	msg, err := scanMessage(row)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: idempotency_key=%s", domain.ErrNotFound, key)
		}
		return nil, fmt.Errorf("postgres: get by idempotency key: %w", err)
	}
	return msg, nil
}


func (s *Store) GetByID(ctx context.Context, id string)(*domain.Message, error){
	const q = `SELECT ` + msgColumns + ` FROM messages where id = @id;`

	row := s.pool.QueryRow(ctx, q, pgx.NamedArgs{"id": id})
	msg, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			return nil, fmt.Errorf("%w: id=%s", domain.ErrNotFound, id)
		}
		return nil, fmt.Errorf("postgres: get by id: %w", err)
	}
	return msg, nil
}