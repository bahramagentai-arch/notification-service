package store
import (
	"context"
	"fmt"
)

func (s *Store)TruncateForTest (ctx context.Context) error{
	const q= `TRUNCATE TABLE messages RESTART IDENTITY CASCADE;`
	_, err := s.pool.Exec(ctx, q)
	if err != nil {
		return fmt.Errorf("postgres: truncate for test: %w", err)
	}
	return nil
}