package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeClient    ErrorType = "client_error"    // 4xx - Client-side errors
	ErrorTypeServer    ErrorType = "server_error"    // 5xx - Server-side errors
	ErrorTypeTransient ErrorType = "transient_error" // Temporary errors (retry recommended)
	ErrorTypePermanent ErrorType = "permanent_error" // Permanent errors (don't retry)
)

// ErrorCode represents specific error conditions
type ErrorCode string

const (
	// Client errors (4xx)
	ErrorCodeBadRequest          ErrorCode = "bad_request"
	ErrorCodeUnauthorized        ErrorCode = "unauthorized"
	ErrorCodeForbidden           ErrorCode = "forbidden"
	ErrorCodeNotFound            ErrorCode = "not_found"
	ErrorCodeConflict            ErrorCode = "conflict"
	ErrorCodeTooLarge            ErrorCode = "entity_too_large"
	ErrorCodeInvalidRange        ErrorCode = "invalid_range"
	ErrorCodePreconditionFailed  ErrorCode = "precondition_failed"

	// Server errors (5xx)
	ErrorCodeInternal            ErrorCode = "internal_error"
	ErrorCodeNotImplemented      ErrorCode = "not_implemented"
	ErrorCodeServiceUnavailable  ErrorCode = "service_unavailable"
	ErrorCodeGatewayTimeout      ErrorCode = "gateway_timeout"

	// Storage errors
	ErrorCodeStorageFailure      ErrorCode = "storage_failure"
	ErrorCodeStorageUnavailable  ErrorCode = "storage_unavailable"
	ErrorCodeStorageQuotaExceeded ErrorCode = "quota_exceeded"

	// Metadata errors
	ErrorCodeMetadataCorrupted   ErrorCode = "metadata_corrupted"
	ErrorCodeMetadataLocked      ErrorCode = "metadata_locked"

	// Network errors
	ErrorCodeNetworkTimeout      ErrorCode = "network_timeout"
	ErrorCodeNetworkUnreachable  ErrorCode = "network_unreachable"
	ErrorCodeConnectionReset     ErrorCode = "connection_reset"

	// Authentication/Authorization errors
	ErrorCodeTokenExpired        ErrorCode = "token_expired"
	ErrorCodeTokenInvalid        ErrorCode = "token_invalid"
	ErrorCodeInsufficientPermissions ErrorCode = "insufficient_permissions"

	// Supply chain errors
	ErrorCodeSignatureInvalid    ErrorCode = "signature_invalid"
	ErrorCodeVerificationFailed  ErrorCode = "verification_failed"
	ErrorCodeSBOMInvalid         ErrorCode = "sbom_invalid"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Type       ErrorType              `json:"type"`
	HTTPStatus int                    `json:"-"`
	Err        error                  `json:"-"`
	Retryable  bool                   `json:"retryable"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
	err.setDefaults()
	return err
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *AppError {
	appErr := &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: make(map[string]interface{}),
	}
	appErr.setDefaults()
	return appErr
}

// WithDetail adds a detail field to the error
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	e.Details[key] = value
	return e
}

// setDefaults sets default values based on error code
func (e *AppError) setDefaults() {
	switch e.Code {
	// Client errors (4xx)
	case ErrorCodeBadRequest:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusBadRequest
		e.Retryable = false
	case ErrorCodeUnauthorized:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusUnauthorized
		e.Retryable = false
	case ErrorCodeForbidden:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusForbidden
		e.Retryable = false
	case ErrorCodeNotFound:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusNotFound
		e.Retryable = false
	case ErrorCodeConflict:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusConflict
		e.Retryable = false
	case ErrorCodeTooLarge:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusRequestEntityTooLarge
		e.Retryable = false
	case ErrorCodeInvalidRange:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusRequestedRangeNotSatisfiable
		e.Retryable = false
	case ErrorCodePreconditionFailed:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusPreconditionFailed
		e.Retryable = false

	// Server errors (5xx) - mostly retryable
	case ErrorCodeInternal:
		e.Type = ErrorTypeServer
		e.HTTPStatus = http.StatusInternalServerError
		e.Retryable = true
	case ErrorCodeNotImplemented:
		e.Type = ErrorTypeServer
		e.HTTPStatus = http.StatusNotImplemented
		e.Retryable = false
	case ErrorCodeServiceUnavailable:
		e.Type = ErrorTypeTransient
		e.HTTPStatus = http.StatusServiceUnavailable
		e.Retryable = true
	case ErrorCodeGatewayTimeout:
		e.Type = ErrorTypeTransient
		e.HTTPStatus = http.StatusGatewayTimeout
		e.Retryable = true

	// Storage errors - some retryable
	case ErrorCodeStorageFailure:
		e.Type = ErrorTypeServer
		e.HTTPStatus = http.StatusInternalServerError
		e.Retryable = true
	case ErrorCodeStorageUnavailable:
		e.Type = ErrorTypeTransient
		e.HTTPStatus = http.StatusServiceUnavailable
		e.Retryable = true
	case ErrorCodeStorageQuotaExceeded:
		e.Type = ErrorTypePermanent
		e.HTTPStatus = http.StatusInsufficientStorage
		e.Retryable = false

	// Metadata errors
	case ErrorCodeMetadataCorrupted:
		e.Type = ErrorTypeServer
		e.HTTPStatus = http.StatusInternalServerError
		e.Retryable = false
	case ErrorCodeMetadataLocked:
		e.Type = ErrorTypeTransient
		e.HTTPStatus = http.StatusConflict
		e.Retryable = true

	// Network errors - all retryable
	case ErrorCodeNetworkTimeout:
		e.Type = ErrorTypeTransient
		e.HTTPStatus = http.StatusGatewayTimeout
		e.Retryable = true
	case ErrorCodeNetworkUnreachable:
		e.Type = ErrorTypeTransient
		e.HTTPStatus = http.StatusServiceUnavailable
		e.Retryable = true
	case ErrorCodeConnectionReset:
		e.Type = ErrorTypeTransient
		e.HTTPStatus = http.StatusServiceUnavailable
		e.Retryable = true

	// Auth errors
	case ErrorCodeTokenExpired:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusUnauthorized
		e.Retryable = false
	case ErrorCodeTokenInvalid:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusUnauthorized
		e.Retryable = false
	case ErrorCodeInsufficientPermissions:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusForbidden
		e.Retryable = false

	// Supply chain errors
	case ErrorCodeSignatureInvalid:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusBadRequest
		e.Retryable = false
	case ErrorCodeVerificationFailed:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusUnprocessableEntity
		e.Retryable = false
	case ErrorCodeSBOMInvalid:
		e.Type = ErrorTypeClient
		e.HTTPStatus = http.StatusBadRequest
		e.Retryable = false

	default:
		e.Type = ErrorTypeServer
		e.HTTPStatus = http.StatusInternalServerError
		e.Retryable = false
	}
}

// IsRetryable checks if the error is retryable
func IsRetryable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Retryable
	}
	return false
}

// IsTransient checks if the error is transient
func IsTransient(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeTransient
	}
	return false
}

// GetHTTPStatus extracts the HTTP status code from an error
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// Common error constructors for convenience

func NewBadRequest(message string) *AppError {
	return New(ErrorCodeBadRequest, message)
}

func NewUnauthorized(message string) *AppError {
	return New(ErrorCodeUnauthorized, message)
}

func NewForbidden(message string) *AppError {
	return New(ErrorCodeForbidden, message)
}

func NewNotFound(message string) *AppError {
	return New(ErrorCodeNotFound, message)
}

func NewConflict(message string) *AppError {
	return New(ErrorCodeConflict, message)
}

func NewInternal(message string) *AppError {
	return New(ErrorCodeInternal, message)
}

func NewServiceUnavailable(message string) *AppError {
	return New(ErrorCodeServiceUnavailable, message)
}
