package reliability_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/reliability"
	"github.com/candlekeep/zot-artifact-store/test"
	"zotregistry.io/zot/pkg/log"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("Starts in closed state", func(t *testing.T) {
		// Given: A new circuit breaker
		logger := log.NewLogger("debug", "")
		cb := reliability.NewCircuitBreaker(reliability.DefaultCircuitBreakerConfig(), logger)

		// When: Checking initial state
		state := cb.GetState()

		// Then: State is closed
		test.AssertEqual(t, reliability.StateClosed, state, "initial state")
	})

	t.Run("Opens after max failures", func(t *testing.T) {
		// Given: A circuit breaker with low threshold
		logger := log.NewLogger("debug", "")
		config := &reliability.CircuitBreakerConfig{
			MaxFailures:     3,
			Timeout:         1 * time.Second,
			HalfOpenSuccess: 2,
			HalfOpenMax:     3,
		}
		cb := reliability.NewCircuitBreaker(config, logger)

		// When: Executing failing function multiple times
		for i := 0; i < 3; i++ {
			cb.Execute(context.Background(), func(ctx context.Context) error {
				return fmt.Errorf("failure")
			})
		}

		// Then: Circuit opens
		test.AssertEqual(t, reliability.StateOpen, cb.GetState(), "should be open")
	})

	t.Run("Rejects requests when open", func(t *testing.T) {
		// Given: An open circuit breaker
		logger := log.NewLogger("debug", "")
		config := &reliability.CircuitBreakerConfig{
			MaxFailures:     2,
			Timeout:         1 * time.Second,
			HalfOpenSuccess: 1,
			HalfOpenMax:     2,
		}
		cb := reliability.NewCircuitBreaker(config, logger)

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(context.Background(), func(ctx context.Context) error {
				return fmt.Errorf("failure")
			})
		}

		// When: Attempting execution
		err := cb.Execute(context.Background(), func(ctx context.Context) error {
			return nil
		})

		// Then: Request is rejected
		test.AssertError(t, err, "should reject request")
	})

	t.Run("Transitions to half-open after timeout", func(t *testing.T) {
		// Given: An open circuit breaker with short timeout
		logger := log.NewLogger("debug", "")
		config := &reliability.CircuitBreakerConfig{
			MaxFailures:     2,
			Timeout:         100 * time.Millisecond,
			HalfOpenSuccess: 1,
			HalfOpenMax:     2,
		}
		cb := reliability.NewCircuitBreaker(config, logger)

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(context.Background(), func(ctx context.Context) error {
				return fmt.Errorf("failure")
			})
		}

		// When: Waiting for timeout
		time.Sleep(150 * time.Millisecond)

		// Execute to trigger state transition
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return nil
		})

		// Then: State is half-open or closed
		state := cb.GetState()
		test.AssertTrue(t, state == reliability.StateHalfOpen || state == reliability.StateClosed, "should transition from open")
	})

	t.Run("Closes after successful half-open attempts", func(t *testing.T) {
		// Given: A half-open circuit breaker
		logger := log.NewLogger("debug", "")
		config := &reliability.CircuitBreakerConfig{
			MaxFailures:     2,
			Timeout:         50 * time.Millisecond,
			HalfOpenSuccess: 2,
			HalfOpenMax:     3,
		}
		cb := reliability.NewCircuitBreaker(config, logger)

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(context.Background(), func(ctx context.Context) error {
				return fmt.Errorf("failure")
			})
		}

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// When: Executing successful requests in half-open
		for i := 0; i < 2; i++ {
			cb.Execute(context.Background(), func(ctx context.Context) error {
				return nil
			})
		}

		// Then: Circuit closes
		test.AssertEqual(t, reliability.StateClosed, cb.GetState(), "should close")
	})

	t.Run("Reopens on failure in half-open state", func(t *testing.T) {
		// Given: A half-open circuit breaker
		logger := log.NewLogger("debug", "")
		config := &reliability.CircuitBreakerConfig{
			MaxFailures:     2,
			Timeout:         50 * time.Millisecond,
			HalfOpenSuccess: 2,
			HalfOpenMax:     3,
		}
		cb := reliability.NewCircuitBreaker(config, logger)

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(context.Background(), func(ctx context.Context) error {
				return fmt.Errorf("failure")
			})
		}

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// When: Failing in half-open state
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return fmt.Errorf("failure")
		})

		// Then: Circuit reopens
		test.AssertEqual(t, reliability.StateOpen, cb.GetState(), "should reopen")
	})

	t.Run("Reset manually closes circuit", func(t *testing.T) {
		// Given: An open circuit breaker
		logger := log.NewLogger("debug", "")
		config := &reliability.CircuitBreakerConfig{
			MaxFailures:     2,
			Timeout:         1 * time.Second,
			HalfOpenSuccess: 1,
			HalfOpenMax:     2,
		}
		cb := reliability.NewCircuitBreaker(config, logger)

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(context.Background(), func(ctx context.Context) error {
				return fmt.Errorf("failure")
			})
		}

		// When: Resetting
		cb.Reset()

		// Then: Circuit is closed
		test.AssertEqual(t, reliability.StateClosed, cb.GetState(), "should be closed")
	})

	t.Run("GetMetrics returns current state", func(t *testing.T) {
		// Given: A circuit breaker
		logger := log.NewLogger("debug", "")
		cb := reliability.NewCircuitBreaker(reliability.DefaultCircuitBreakerConfig(), logger)

		// When: Getting metrics
		metrics := cb.GetMetrics()

		// Then: Metrics contain state
		test.AssertTrue(t, metrics["state"] != nil, "has state")
	})
}

func TestCircuitBreakerManager(t *testing.T) {
	t.Run("Creates circuit breakers by name", func(t *testing.T) {
		// Given: A circuit breaker manager
		logger := log.NewLogger("debug", "")
		manager := reliability.NewCircuitBreakerManager(nil, logger)

		// When: Getting breakers by name
		cb1 := manager.GetBreaker("service1")
		cb2 := manager.GetBreaker("service2")
		cb1Again := manager.GetBreaker("service1")

		// Then: Returns same instance for same name
		test.AssertTrue(t, cb1 != nil, "creates breaker")
		test.AssertTrue(t, cb2 != nil, "creates different breaker")
		test.AssertTrue(t, cb1 == cb1Again, "returns same instance")
	})

	t.Run("Execute through manager", func(t *testing.T) {
		// Given: A circuit breaker manager
		logger := log.NewLogger("debug", "")
		manager := reliability.NewCircuitBreakerManager(nil, logger)

		// When: Executing through manager
		err := manager.Execute(context.Background(), "test", func(ctx context.Context) error {
			return nil
		})

		// Then: Succeeds
		test.AssertNoError(t, err, "should succeed")
	})

	t.Run("GetAllMetrics returns metrics for all breakers", func(t *testing.T) {
		// Given: A manager with multiple breakers
		logger := log.NewLogger("debug", "")
		manager := reliability.NewCircuitBreakerManager(nil, logger)

		manager.GetBreaker("service1")
		manager.GetBreaker("service2")

		// When: Getting all metrics
		metrics := manager.GetAllMetrics()

		// Then: Contains metrics for both
		test.AssertTrue(t, len(metrics) == 2, "has metrics for both breakers")
	})
}
