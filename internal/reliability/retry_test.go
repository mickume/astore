package reliability_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/errors"
	"github.com/candlekeep/zot-artifact-store/internal/reliability"
	"github.com/candlekeep/zot-artifact-store/test"
	"zotregistry.io/zot/pkg/log"
)

func TestRetryMechanism(t *testing.T) {
	t.Run("Succeeds on first attempt", func(t *testing.T) {
		// Given: A retryer and a function that succeeds
		logger := log.NewLogger("debug", "")
		retryer := reliability.NewRetryer(reliability.DefaultRetryPolicy(), logger)

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return nil // Success
		}

		// When: Executing the function
		err := retryer.Do(context.Background(), fn)

		// Then: Succeeds without retry
		test.AssertNoError(t, err, "should succeed")
		test.AssertEqual(t, 1, attempts, "should only attempt once")
	})

	t.Run("Retries on retryable error", func(t *testing.T) {
		// Given: A retryer and a function that fails then succeeds
		logger := log.NewLogger("debug", "")
		policy := &reliability.RetryPolicy{
			MaxAttempts:     3,
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        100 * time.Millisecond,
			Multiplier:      2.0,
			RandomizeFactor: 0,
		}
		retryer := reliability.NewRetryer(policy, logger)

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return errors.NewServiceUnavailable("service down")
			}
			return nil
		}

		// When: Executing the function
		err := retryer.Do(context.Background(), fn)

		// Then: Succeeds after retries
		test.AssertNoError(t, err, "should succeed after retries")
		test.AssertEqual(t, 3, attempts, "should attempt 3 times")
	})

	t.Run("Does not retry on non-retryable error", func(t *testing.T) {
		// Given: A retryer and a function that returns non-retryable error
		logger := log.NewLogger("debug", "")
		retryer := reliability.NewRetryer(reliability.DefaultRetryPolicy(), logger)

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return errors.NewNotFound("not found")
		}

		// When: Executing the function
		err := retryer.Do(context.Background(), fn)

		// Then: Fails immediately without retry
		test.AssertError(t, err, "should fail")
		test.AssertEqual(t, 1, attempts, "should only attempt once")
	})

	t.Run("Fails after max attempts", func(t *testing.T) {
		// Given: A retryer and a function that always fails
		logger := log.NewLogger("debug", "")
		policy := &reliability.RetryPolicy{
			MaxAttempts:     3,
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        100 * time.Millisecond,
			Multiplier:      2.0,
			RandomizeFactor: 0,
		}
		retryer := reliability.NewRetryer(policy, logger)

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return errors.NewServiceUnavailable("always fails")
		}

		// When: Executing the function
		err := retryer.Do(context.Background(), fn)

		// Then: Fails after max attempts
		test.AssertError(t, err, "should fail")
		test.AssertEqual(t, 3, attempts, "should attempt max times")
	})

	t.Run("Respects context cancellation", func(t *testing.T) {
		// Given: A retryer and a context that will be cancelled
		logger := log.NewLogger("debug", "")
		policy := &reliability.RetryPolicy{
			MaxAttempts:     5,
			InitialDelay:    100 * time.Millisecond,
			MaxDelay:        1 * time.Second,
			Multiplier:      2.0,
			RandomizeFactor: 0,
		}
		retryer := reliability.NewRetryer(policy, logger)

		ctx, cancel := context.WithCancel(context.Background())

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			if attempts == 1 {
				// Cancel after first attempt
				cancel()
			}
			return errors.NewServiceUnavailable("service down")
		}

		// When: Executing the function
		err := retryer.Do(ctx, fn)

		// Then: Stops on context cancellation
		test.AssertError(t, err, "should fail due to cancellation")
		test.AssertTrue(t, attempts <= 2, "should stop early")
	})
}

func TestRetryWithCallback(t *testing.T) {
	t.Run("Callback is called for each attempt", func(t *testing.T) {
		// Given: A retryer with callback
		logger := log.NewLogger("debug", "")
		policy := &reliability.RetryPolicy{
			MaxAttempts:     3,
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        100 * time.Millisecond,
			Multiplier:      2.0,
			RandomizeFactor: 0,
		}
		retryer := reliability.NewRetryer(policy, logger)

		attempts := 0
		callbackCalls := 0

		fn := func(ctx context.Context) error {
			attempts++
			if attempts < 2 {
				return errors.NewServiceUnavailable("service down")
			}
			return nil
		}

		callback := func(attempt int, err error) {
			callbackCalls++
		}

		// When: Executing with callback
		err := retryer.DoWithCallback(context.Background(), fn, callback)

		// Then: Callback called for each attempt
		test.AssertNoError(t, err, "should succeed")
		test.AssertEqual(t, 2, callbackCalls, "callback called for each attempt")
	})
}

func TestRetryPolicies(t *testing.T) {
	t.Run("Default policy has sensible values", func(t *testing.T) {
		// Given: Default policy
		policy := reliability.DefaultRetryPolicy()

		// Then: Has expected values
		test.AssertEqual(t, 3, policy.MaxAttempts, "max attempts")
		test.AssertTrue(t, policy.InitialDelay > 0, "has initial delay")
		test.AssertTrue(t, policy.Multiplier > 1, "has multiplier")
	})

	t.Run("Aggressive policy has more attempts", func(t *testing.T) {
		// Given: Aggressive policy
		policy := reliability.AggressiveRetryPolicy()

		// Then: Has more attempts than default
		test.AssertTrue(t, policy.MaxAttempts >= 5, "has at least 5 attempts")
	})

	t.Run("Conservative policy has fewer attempts", func(t *testing.T) {
		// Given: Conservative policy
		policy := reliability.ConservativeRetryPolicy()

		// Then: Has fewer attempts than default
		test.AssertTrue(t, policy.MaxAttempts <= 2, "has at most 2 attempts")
	})
}

func TestShouldRetry(t *testing.T) {
	t.Run("Returns true for retryable errors", func(t *testing.T) {
		// Given: A retryable error
		err := errors.NewServiceUnavailable("down")

		// When: Checking if should retry
		shouldRetry := reliability.ShouldRetry(err)

		// Then: Returns true
		test.AssertTrue(t, shouldRetry, "should retry")
	})

	t.Run("Returns false for non-retryable errors", func(t *testing.T) {
		// Given: A non-retryable error
		err := errors.NewNotFound("not found")

		// When: Checking if should retry
		shouldRetry := reliability.ShouldRetry(err)

		// Then: Returns false
		test.AssertFalse(t, shouldRetry, "should not retry")
	})

	t.Run("Returns false for nil error", func(t *testing.T) {
		// When: Checking nil error
		shouldRetry := reliability.ShouldRetry(nil)

		// Then: Returns false
		test.AssertFalse(t, shouldRetry, "nil error should not retry")
	})

	t.Run("Returns false for standard errors", func(t *testing.T) {
		// Given: A standard error
		err := fmt.Errorf("standard error")

		// When: Checking if should retry
		shouldRetry := reliability.ShouldRetry(err)

		// Then: Returns false
		test.AssertFalse(t, shouldRetry, "standard error should not retry")
	})
}
