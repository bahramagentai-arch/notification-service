package domain

import "errors"


var (
	ErrDuplicate = errors.New("notification: duplicate idempotency key")
	EmptyBody = errors.New("notification: empty body")
	ToLongBodyLen = errors.New("notification: to long body len")
	ErrInvalidRecipient = errors.New("notification: invalid recipient")
	ErrNotFound = errors.New("notification: not found")
)