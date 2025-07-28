package apperror

import (
	"fmt"
	"net/http"
)

// --- User Error ---
type UserError struct {
	StatusCode int
	Message    string
}

func (e *UserError) Error() string {
	return fmt.Sprintf("UserError (code: %d): %s", e.StatusCode, e.Message)
}

func NewUserError(action, details string) *UserError {
	msg := fmt.Sprintf("user error during %s: %s", action, details)
	return &UserError{
		StatusCode: http.StatusBadRequest,
		Message:    msg,
	}
}

// --- Contact Error ---
type ContactError struct {
	StatusCode int
	Message    string
}

func (e *ContactError) Error() string {
	return fmt.Sprintf("ContactError (code: %d): %s", e.StatusCode, e.Message)
}

func NewContactError(operation, reason string) *ContactError {
	msg := fmt.Sprintf("contact error during %s: %s", operation, reason)
	return &ContactError{
		StatusCode: http.StatusConflict,
		Message:    msg,
	}
}

// --- Contact Detail Error ---
type ContactDetailError struct {
	StatusCode int
	Message    string
}

func (e *ContactDetailError) Error() string {
	return fmt.Sprintf("ContactDetailError (code: %d): %s", e.StatusCode, e.Message)
}

func NewContactDetailError(process, problem string) *ContactDetailError {
	msg := fmt.Sprintf("contact detail error during %s: %s", process, problem)
	return &ContactDetailError{
		StatusCode: http.StatusUnprocessableEntity,
		Message:    msg,
	}
}

// --- Authorization Error ---
type AuthError struct {
	StatusCode int
	Message    string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("Unauthorized (code: %d): %s", e.StatusCode, e.Message)
}

func NewAuthError(action string) *AuthError {
	msg := fmt.Sprintf("unauthorized access while trying to %s", action)
	return &AuthError{
		StatusCode: http.StatusUnauthorized,
		Message:    msg,
	}
}

// --- Not Found Error ---
type NotFoundError struct {
	StatusCode int
	Message    string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("NotFound (code: %d): %s", e.StatusCode, e.Message)
}

func NewNotFoundError(resource string, id int) *NotFoundError {
	msg := fmt.Sprintf("%s with ID %d not found", resource, id)
	return &NotFoundError{
		StatusCode: http.StatusNotFound,
		Message:    msg,
	}
}

// --- Validation Error ---
type ValidationError struct {
	StatusCode int
	Message    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("ValidationError (code: %d): %s", e.StatusCode, e.Message)
}

func NewValidationError(fieldName, message string) *ValidationError {
	msg := fmt.Sprintf("invalid %s: %s", fieldName, message)
	return &ValidationError{
		StatusCode: http.StatusBadRequest,
		Message:    msg,
	}
}
