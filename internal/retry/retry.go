package retry

import (
	"context"
	"time"
)

// Do executes a function with exponential backoff retries.
func Do(ctx context.Context, maxRetries int, baseDelay time.Duration, fn func() error) error {
	var err error
	delay := baseDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
			delay *= 2 // Exponential backoff
		}

		err = fn()
		if err == nil {
			return nil
		}
	}
	return err
}
