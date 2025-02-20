package errors

import "errors"

// Common error types
var (
	ErrInvalidCredentials = errors.New("Invalid email or password")
	ErrEmailExists        = errors.New("An account with this email already exists")
	ErrUsernameExists     = errors.New("This username is already taken")
	ErrUserNotFound       = errors.New("User not found")
	ErrInvalidUserID      = errors.New("Invalid user ID")
	ErrInvalidPassword    = errors.New("Current password is incorrect")
	ErrWeakPassword       = errors.New("Password must be at least 8 characters long")
	ErrInvalidEmail       = errors.New("Please enter a valid email address")
	ErrInvalidInput       = errors.New("Please check your input and try again")
	ErrUnauthorized       = errors.New("Please log in to continue")
	ErrServerError        = errors.New("Something went wrong. Please try again later")
)

// ValidationError represents a validation error with a user-friendly message
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors holds multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	return ve[0].Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(ValidationErrors)
	return ok
}

// FormatValidationErrors formats validation errors into a map
func FormatValidationErrors(errs ValidationErrors) map[string]string {
	errorMap := make(map[string]string)
	for _, err := range errs {
		errorMap[err.Field] = err.Message
	}
	return errorMap
}
