package domain

import "errors"


var ErrDuplicate = errors.New("notification: duplicate idempotency key")