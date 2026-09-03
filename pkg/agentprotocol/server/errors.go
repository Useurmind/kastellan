package server

import (
	"fmt"
	"time"
)

// ProtocolError represents a protocol-level error.
type ProtocolError struct {
	Code    string
	Message string
}

// Error returns the error message.
func (e *ProtocolError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Is returns whether this error matches the target error.
func (e *ProtocolError) Is(target error) bool {
	te, ok := target.(*ProtocolError)
	if !ok {
		return false
	}
	return e.Code == te.Code
}

// ConnectionError represents a connection error.
type ConnectionError struct {
	Err       error
	Retryable bool
}

// Error returns the error message.
func (e *ConnectionError) Error() string {
	if e.Retryable {
		return fmt.Sprintf("connection error (retryable): %v", e.Err)
	}
	return fmt.Sprintf("connection error: %v", e.Err)
}

// Unwrap returns the underlying error.
func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// IsRetryable returns whether the error is retryable.
func (e *ConnectionError) IsRetryable() bool {
	return e.Retryable
}

// AuthenticationError represents an authentication error.
type AuthenticationError struct {
	Message string
}

// Error returns the error message.
func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.Message)
}

// AuthorizationError represents an authorization error.
type AuthorizationError struct {
	Message string
}

// Error returns the error message.
func (e *AuthorizationError) Error() string {
	return fmt.Sprintf("authorization failed: %s", e.Message)
}

// TimeoutError represents a timeout error.
type TimeoutError struct {
	Duration time.Duration
	Message  string
}

// Error returns the error message.
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout after %v: %s", e.Duration, e.Message)
}

// NetworkError represents a network error.
type NetworkError struct {
	Err error
}

// Error returns the error message.
func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

// Unwrap returns the underlying error.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// ResourceError represents a resource-related error.
type ResourceError struct {
	Resource string
	Code     string
	Message  string
}

// Error returns the error message.
func (e *ResourceError) Error() string {
	return fmt.Sprintf("resource %s error (%s): %s", e.Resource, e.Code, e.Message)
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error (%s): %s", e.Field, e.Message)
}

// NewProtocolError creates a new protocol error.
func NewProtocolError(code, message string) *ProtocolError {
	return &ProtocolError{
		Code:    code,
		Message: message,
	}
}

// NewConnectionError creates a new connection error.
func NewConnectionError(err error, retryable bool) *ConnectionError {
	return &ConnectionError{
		Err:       err,
		Retryable: retryable,
	}
}

// NewAuthenticationError creates a new authentication error.
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		Message: message,
	}
}

// NewAuthorizationError creates a new authorization error.
func NewAuthorizationError(message string) *AuthorizationError {
	return &AuthorizationError{
		Message: message,
	}
}

// NewTimeoutError creates a new timeout error.
func NewTimeoutError(duration time.Duration, message string) *TimeoutError {
	return &TimeoutError{
		Duration: duration,
		Message:  message,
	}
}

// NewNetworkError creates a new network error.
func NewNetworkError(err error) *NetworkError {
	return &NetworkError{
		Err: err,
	}
}

// NewResourceError creates a new resource error.
func NewResourceError(resource, code, message string) *ResourceError {
	return &ResourceError{
		Resource: resource,
		Code:     code,
		Message:  message,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
