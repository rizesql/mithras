package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Policy struct {
	attempts    int
	backoff     BackoffFunc
	shouldRetry func(error) bool
}

func New(opts ...Option) *Policy {
	c := &Policy{
		attempts:    3,
		backoff:     LinBackoff(100 * time.Millisecond),
		shouldRetry: func(error) bool { return true },
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option func(c *Policy)

func Attempts(attempts int) Option {
	return func(c *Policy) { c.attempts = attempts }
}

func Backoff(backoff BackoffFunc) Option {
	return func(c *Policy) { c.backoff = backoff }
}

func ShouldRetry(shouldRetry func(error) bool) Option {
	return func(c *Policy) { c.shouldRetry = shouldRetry }
}

func DoResult[T any](ctx context.Context, p *Policy, fn func(context.Context) (T, error)) (t T, err error) {
	if p == nil {
		p = New()
	}

	if p.attempts < 1 {
		return t, fmt.Errorf("attempts must be greater than 0")
	}

	for i := 1; i <= p.attempts; i++ {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return t, ctxErr
		}

		t, err = fn(ctx)
		if err == nil {
			return t, nil
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return t, err
		}

		if !p.shouldRetry(err) {
			return t, err
		}

		if i == p.attempts {
			break
		}

		delay := p.backoff(i)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return t, ctx.Err()
		case <-timer.C:
		}
	}

	return t, err
}

func Do(ctx context.Context, p *Policy, fn func(context.Context) error) error {
	_, err := DoResult(ctx, p, func(ctx context.Context) (any, error) {
		return nil, fn(ctx)
	})
	return err
}

// func Do(ctx context.Context, p *Policy, fn func(context.Context) error) (err error) {
// 	if p == nil {
// 		p = New()
// 	}

// 	if p.attempts < 1 {
// 		return fmt.Errorf("attempts must be greater than 0")
// 	}

// 	for i := 1; i <= p.attempts; i++ {
// 		if ctxErr := ctx.Err(); ctxErr != nil {
// 			return ctxErr
// 		}

// 		err = fn(ctx)
// 		if err == nil {
// 			return nil
// 		}

// 		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
// 			return err
// 		}

// 		if !p.shouldRetry(err) {
// 			return err
// 		}

// 		if i == p.attempts {
// 			break
// 		}

// 		delay := p.backoff(i)

// 		if p.sleep != nil {
// 			p.sleep(delay)
// 			continue
// 		}

// 		timer := time.NewTimer(delay)
// 		select {
// 		case <-ctx.Done():
// 			if !timer.Stop() {
// 				<-timer.C
// 			}
// 			return ctx.Err()
// 		case <-timer.C:
// 		}
// 	}

// 	return err
// }
