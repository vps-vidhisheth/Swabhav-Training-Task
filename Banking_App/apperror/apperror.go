package apperror

import (
	"fmt"
	"net/http"
)

// --- Bank Error ---

type BankError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *BankError) Error() string {
	return fmt.Sprintf("%s (code: %d): %s", e.Err.Error(), e.StatusCode, e.Message)
}

func NewBankError(action, reason string) *BankError {
	msg := fmt.Sprintf("bank error during %s", action)
	return &BankError{
		Err:        fmt.Errorf("BankError"),
		StatusCode: http.StatusConflict,
		Message:    fmt.Sprintf("%s: %s", msg, reason),
	}
}

// --- Customer Error ---

type CustomerError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *CustomerError) Error() string {
	return fmt.Sprintf("%s (code: %d): %s", e.Err.Error(), e.StatusCode, e.Message)
}

func NewCustomerError(action, reason string) *CustomerError {
	msg := fmt.Sprintf("customer error during %s", action)
	return &CustomerError{
		Err:        fmt.Errorf("CustomerError"),
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf("%s: %s", msg, reason),
	}
}

// --- Account Error ---

type AccountError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *AccountError) Error() string {
	return fmt.Sprintf("%s (code: %d): %s", e.Err.Error(), e.StatusCode, e.Message)
}

func NewAccountError(action, reason string) *AccountError {
	msg := fmt.Sprintf("account error during %s", action)
	return &AccountError{
		Err:        fmt.Errorf("AccountError"),
		StatusCode: http.StatusUnprocessableEntity,
		Message:    fmt.Sprintf("%s: %s", msg, reason),
	}
}

// --- Not Found Error ---

type NotFoundError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s (code: %d): %s", e.Err.Error(), e.StatusCode, e.Message)
}

func NewNotFoundError(resource string, id int) *NotFoundError {
	return &NotFoundError{
		Err:        fmt.Errorf("NotFound"),
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("%s with ID %d not found", resource, id),
	}
}

// --- Validation Error ---

type ValidationError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s (code: %d): %s", e.Err.Error(), e.StatusCode, e.Message)
}

func NewValidationError(fieldName, message string) *ValidationError {
	return &ValidationError{
		Err:        fmt.Errorf("ValidationError"),
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf("invalid %s: %s", fieldName, message),
	}
}

// --- Authorization Error ---

type AuthError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("%s (code: %d): %s", e.Err.Error(), e.StatusCode, e.Message)
}

func NewAuthError(action string) *AuthError {
	return &AuthError{
		Err:        fmt.Errorf("Unauthorized"),
		StatusCode: http.StatusUnauthorized,
		Message:    fmt.Sprintf("unauthorized access while trying to %s", action),
	}
}

// --- User Error ---

type UserError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *UserError) Error() string {
	return fmt.Sprintf("%s (code: %d): %s", e.Err.Error(), e.StatusCode, e.Message)
}

func NewUserError(action, reason string) *UserError {
	msg := fmt.Sprintf("user error during %s", action)
	return &UserError{
		Err:        fmt.Errorf("UserError"),
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf("%s: %s", msg, reason),
	}
}
