package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/BahramRousta/notification-service/internal/notification/domain"
)


func validateBody(body string) error {
	if strings.TrimSpace(body) == ""{
		return fmt.Errorf("%w: body is empty", domain.EmptyBody)
	}

	n := utf8.RuneCountInString(body)
	if n > maxBodyRunes {
		return fmt.Errorf("%w: body has %d runes, max is %d", domain.ToLongBodyLen, n, maxBodyRunes)
	}
	return nil
}

func validateRecipient(r string) error {
	r = strings.TrimSpace(r)
	if r == "" {
		return fmt.Errorf("%w: recipient is empty", domain.ErrInvalidRecipient)
	}
	if r[0] != '+' {
		return fmt.Errorf("%w: must start with '+' (E.164 format)", domain.ErrInvalidRecipient)
	}
	if len(r) < 8 || len(r) > 16 {
		return fmt.Errorf("%w: length %d out of range [8,16]", domain.ErrInvalidRecipient, len(r))
	}
	for _, c := range r[1:] {
		if c < '0' || c > '9' {
			return fmt.Errorf("%w: non-digit character '%c'", domain.ErrInvalidRecipient, c)
		}
	}
	return nil
}


func validateRequest (req SendRequest) error {
	if err := validateRecipient(req.Recipient); err != nil {
		return err
	}
	if err := validateBody(req.Body); err != nil {
		return err
	}
	return nil
}