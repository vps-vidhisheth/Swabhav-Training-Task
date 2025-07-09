package apperror

import (
	"fmt"
	"net/http"
)

type BankError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *BankError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("BankError (code: %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("BankError (code: %d): %s", e.StatusCode, e.Message)
}

func (e *BankError) Unwrap() error {
	return e.Err
}

func NewBankError(action, reason string, cause ...error) *BankError {
	msg := fmt.Sprintf("bank error during %s", action)
	var errCause error
	if len(cause) > 0 {
		errCause = cause[0]
	}
	return &BankError{
		Err:        errCause,
		StatusCode: http.StatusConflict,
		Message:    fmt.Sprintf("%s: %s", msg, reason),
	}
}

type CustomerError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *CustomerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("CustomerError (code: %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("CustomerError (code: %d): %s", e.StatusCode, e.Message)
}

func (e *CustomerError) Unwrap() error {
	return e.Err
}

func NewCustomerError(action, reason string, cause ...error) *CustomerError {
	msg := fmt.Sprintf("customer error during %s", action)
	var errCause error
	if len(cause) > 0 {
		errCause = cause[0]
	}
	return &CustomerError{
		Err:        errCause,
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf("%s: %s", msg, reason),
	}
}

type AccountError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *AccountError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("AccountError (code: %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("AccountError (code: %d): %s", e.StatusCode, e.Message)
}

func (e *AccountError) Unwrap() error {
	return e.Err
}

func NewAccountError(action, reason string, cause ...error) *AccountError {
	msg := fmt.Sprintf("account error during %s", action)
	var errCause error
	if len(cause) > 0 {
		errCause = cause[0]
	}
	return &AccountError{
		Err:        errCause,
		StatusCode: http.StatusUnprocessableEntity,
		Message:    fmt.Sprintf("%s: %s", msg, reason),
	}
}

type NotFoundError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *NotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("NotFoundError (code: %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("NotFoundError (code: %d): %s", e.StatusCode, e.Message)
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(resource string, id int, cause ...error) *NotFoundError {
	var errCause error
	if len(cause) > 0 {
		errCause = cause[0]
	}
	return &NotFoundError{
		Err:        errCause,
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("%s with ID %d not found", resource, id),
	}
}

type ValidationError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("ValidationError (code: %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("ValidationError (code: %d): %s", e.StatusCode, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func NewValidationError(fieldName, message string, cause ...error) *ValidationError {
	var errCause error
	if len(cause) > 0 {
		errCause = cause[0]
	}
	return &ValidationError{
		Err:        errCause,
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf("invalid %s: %s", fieldName, message),
	}
}

type AuthError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *AuthError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("AuthError (code: %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("AuthError (code: %d): %s", e.StatusCode, e.Message)
}

func (e *AuthError) Unwrap() error {
	return e.Err
}

func NewAuthError(action string, cause ...error) *AuthError {
	var errCause error
	if len(cause) > 0 {
		errCause = cause[0]
	}
	return &AuthError{
		Err:        errCause,
		StatusCode: http.StatusUnauthorized,
		Message:    fmt.Sprintf("unauthorized access while trying to %s", action),
	}
}

type UserError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *UserError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("UserError (code: %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("UserError (code: %d): %s", e.StatusCode, e.Message)
}

func (e *UserError) Unwrap() error {
	return e.Err
}

func NewUserError(action, reason string, cause ...error) *UserError {
	msg := fmt.Sprintf("user error during %s", action)
	var errCause error
	if len(cause) > 0 {
		errCause = cause[0]
	}
	return &UserError{
		Err:        errCause,
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf("%s: %s", msg, reason),
	}
}
