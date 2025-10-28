package errors_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/candlekeep/zot-artifact-store/internal/errors"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestErrorClassification(t *testing.T) {
	t.Run("Bad request error is client error", func(t *testing.T) {
		// Given: A bad request error
		err := errors.NewBadRequest("invalid input")

		// When: Checking error properties
		// Then: Error is classified correctly
		test.AssertEqual(t, errors.ErrorCodeBadRequest, err.Code, "error code")
		test.AssertEqual(t, errors.ErrorTypeClient, err.Type, "error type")
		test.AssertEqual(t, http.StatusBadRequest, err.HTTPStatus, "HTTP status")
		test.AssertFalse(t, err.Retryable, "should not be retryable")
	})

	t.Run("Service unavailable is transient and retryable", func(t *testing.T) {
		// Given: A service unavailable error
		err := errors.NewServiceUnavailable("service temporarily unavailable")

		// When: Checking error properties
		// Then: Error is transient and retryable
		test.AssertEqual(t, errors.ErrorCodeServiceUnavailable, err.Code, "error code")
		test.AssertEqual(t, errors.ErrorTypeTransient, err.Type, "error type")
		test.AssertEqual(t, http.StatusServiceUnavailable, err.HTTPStatus, "HTTP status")
		test.AssertTrue(t, err.Retryable, "should be retryable")
	})

	t.Run("Internal error is retryable", func(t *testing.T) {
		// Given: An internal server error
		err := errors.NewInternal("database connection failed")

		// When: Checking error properties
		// Then: Error is server error and retryable
		test.AssertEqual(t, errors.ErrorCodeInternal, err.Code, "error code")
		test.AssertEqual(t, errors.ErrorTypeServer, err.Type, "error type")
		test.AssertEqual(t, http.StatusInternalServerError, err.HTTPStatus, "HTTP status")
		test.AssertTrue(t, err.Retryable, "should be retryable")
	})

	t.Run("Not found error is not retryable", func(t *testing.T) {
		// Given: A not found error
		err := errors.NewNotFound("resource does not exist")

		// When: Checking error properties
		// Then: Error is not retryable
		test.AssertEqual(t, errors.ErrorCodeNotFound, err.Code, "error code")
		test.AssertFalse(t, err.Retryable, "should not be retryable")
	})
}

func TestErrorWrapping(t *testing.T) {
	t.Run("Wrap existing error with context", func(t *testing.T) {
		// Given: An underlying error
		underlying := fmt.Errorf("connection timeout")

		// When: Wrapping with app error
		err := errors.Wrap(underlying, errors.ErrorCodeNetworkTimeout, "failed to connect to storage")

		// Then: Error preserves underlying error
		test.AssertEqual(t, errors.ErrorCodeNetworkTimeout, err.Code, "error code")
		test.AssertTrue(t, err.Err != nil, "underlying error preserved")
		test.AssertTrue(t, err.Retryable, "network timeout should be retryable")
	})
}

func TestErrorDetails(t *testing.T) {
	t.Run("Add details to error", func(t *testing.T) {
		// Given: An error with details
		err := errors.NewBadRequest("validation failed").
			WithDetail("field", "email").
			WithDetail("reason", "invalid format")

		// When: Checking details
		// Then: Details are present
		test.AssertEqual(t, "email", err.Details["field"], "field detail")
		test.AssertEqual(t, "invalid format", err.Details["reason"], "reason detail")
	})
}

func TestRetryableCheck(t *testing.T) {
	t.Run("IsRetryable returns true for retryable errors", func(t *testing.T) {
		// Given: A retryable error
		err := errors.NewServiceUnavailable("service down")

		// When: Checking if retryable
		retryable := errors.IsRetryable(err)

		// Then: Returns true
		test.AssertTrue(t, retryable, "should be retryable")
	})

	t.Run("IsRetryable returns false for non-retryable errors", func(t *testing.T) {
		// Given: A non-retryable error
		err := errors.NewNotFound("not found")

		// When: Checking if retryable
		retryable := errors.IsRetryable(err)

		// Then: Returns false
		test.AssertFalse(t, retryable, "should not be retryable")
	})

	t.Run("IsRetryable returns false for standard errors", func(t *testing.T) {
		// Given: A standard Go error
		err := fmt.Errorf("some error")

		// When: Checking if retryable
		retryable := errors.IsRetryable(err)

		// Then: Returns false
		test.AssertFalse(t, retryable, "standard error not retryable")
	})
}

func TestTransientCheck(t *testing.T) {
	t.Run("IsTransient returns true for transient errors", func(t *testing.T) {
		// Given: A transient error
		err := errors.New(errors.ErrorCodeNetworkTimeout, "timeout")

		// When: Checking if transient
		transient := errors.IsTransient(err)

		// Then: Returns true
		test.AssertTrue(t, transient, "should be transient")
	})
}

func TestHTTPStatus(t *testing.T) {
	t.Run("GetHTTPStatus returns correct status", func(t *testing.T) {
		// Given: An error with specific HTTP status
		err := errors.NewForbidden("access denied")

		// When: Getting HTTP status
		status := errors.GetHTTPStatus(err)

		// Then: Returns correct status
		test.AssertEqual(t, http.StatusForbidden, status, "HTTP status")
	})

	t.Run("GetHTTPStatus returns 500 for standard errors", func(t *testing.T) {
		// Given: A standard Go error
		err := fmt.Errorf("some error")

		// When: Getting HTTP status
		status := errors.GetHTTPStatus(err)

		// Then: Returns 500
		test.AssertEqual(t, http.StatusInternalServerError, status, "default to 500")
	})
}
