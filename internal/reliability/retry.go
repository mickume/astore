package reliability

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/errors"
	"zotregistry.io/zot/pkg/log"
)

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts     int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay before first retry
	MaxDelay        time.Duration // Maximum delay between retries
	Multiplier      float64       // Backoff multiplier (e.g., 2.0 for exponential)
	RandomizeFactor float64       // Jitter factor (0.0 to 1.0)
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:     3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        10 * time.Second,
		Multiplier:      2.0,
		RandomizeFactor: 0.2,
	}
}

// AggressiveRetryPolicy for critical operations
func AggressiveRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:     5,
		InitialDelay:    50 * time.Millisecond,
		MaxDelay:        30 * time.Second,
		Multiplier:      2.0,
		RandomizeFactor: 0.3,
	}
}

// ConservativeRetryPolicy for non-critical operations
func ConservativeRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:     2,
		InitialDelay:    500 * time.Millisecond,
		MaxDelay:        5 * time.Second,
		Multiplier:      1.5,
		RandomizeFactor: 0.1,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// Retryer handles retry logic with exponential backoff
type Retryer struct {
	policy *RetryPolicy
	logger log.Logger
}

// NewRetryer creates a new retryer with the given policy
func NewRetryer(policy *RetryPolicy, logger log.Logger) *Retryer {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}
	return &Retryer{
		policy: policy,
		logger: logger,
	}
}

// Do executes the function with retry logic
func (r *Retryer) Do(ctx context.Context, fn RetryableFunc) error {
	var lastErr error

	for attempt := 0; attempt < r.policy.MaxAttempts; attempt++ {
		// Execute the function
		err := fn(ctx)

		// Success
		if err == nil {
			if attempt > 0 {
				r.logger.Info().
					Int("attempt", attempt+1).
					Msg("operation succeeded after retry")
			}
			return nil
		}

		// Store error
		lastErr = err

		// Check if error is retryable
		if !errors.IsRetryable(err) {
			r.logger.Debug().
				Err(err).
				Int("attempt", attempt+1).
				Msg("non-retryable error encountered")
			return err
		}

		// Check if we have more attempts
		if attempt+1 >= r.policy.MaxAttempts {
			r.logger.Warn().
				Err(err).
				Int("attempts", attempt+1).
				Msg("max retry attempts reached")
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := r.calculateDelay(attempt)

		r.logger.Debug().
			Err(err).
			Int("attempt", attempt+1).
			Dur("delay", delay).
			Msg("retrying operation")

		// Wait before retrying (with context cancellation support)
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", r.policy.MaxAttempts, lastErr)
}

// DoWithCallback executes the function with retry logic and a callback for each attempt
func (r *Retryer) DoWithCallback(ctx context.Context, fn RetryableFunc, callback func(attempt int, err error)) error {
	var lastErr error

	for attempt := 0; attempt < r.policy.MaxAttempts; attempt++ {
		err := fn(ctx)

		// Call callback
		if callback != nil {
			callback(attempt+1, err)
		}

		if err == nil {
			return nil
		}

		lastErr = err

		if !errors.IsRetryable(err) {
			return err
		}

		if attempt+1 >= r.policy.MaxAttempts {
			break
		}

		delay := r.calculateDelay(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", r.policy.MaxAttempts, lastErr)
}

// calculateDelay calculates the delay for the given attempt with exponential backoff and jitter
func (r *Retryer) calculateDelay(attempt int) time.Duration {
	// Calculate base delay with exponential backoff
	delay := float64(r.policy.InitialDelay) * math.Pow(r.policy.Multiplier, float64(attempt))

	// Apply max delay cap
	if delay > float64(r.policy.MaxDelay) {
		delay = float64(r.policy.MaxDelay)
	}

	// Apply jitter (randomize by +/- RandomizeFactor)
	if r.policy.RandomizeFactor > 0 {
		jitter := delay * r.policy.RandomizeFactor
		delay = delay - jitter + (rand.Float64() * 2 * jitter)
	}

	return time.Duration(delay)
}

// ShouldRetry checks if an error should be retried based on its type
func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}
	return errors.IsRetryable(err)
}

// IsTransientError checks if an error is transient
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}
	return errors.IsTransient(err)
}
