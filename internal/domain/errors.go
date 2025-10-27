package domain

import (
	"fmt"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeIO represents I/O errors (file operations)
	ErrorTypeIO ErrorType = "io"
	// ErrorTypeProvider represents AI provider errors
	ErrorTypeProvider ErrorType = "provider"
	// ErrorTypeTranslation represents translation errors
	ErrorTypeTranslation ErrorType = "translation"
	// ErrorTypeFormat represents format protection errors
	ErrorTypeFormat ErrorType = "format"
	// ErrorTypeTerminology represents terminology errors
	ErrorTypeTerminology ErrorType = "terminology"
	// ErrorTypeConfig represents configuration errors
	ErrorTypeConfig ErrorType = "config"
)

// Error represents a domain error with additional context
type Error struct {
	Type    ErrorType
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *Error) Error() string {
	var msg string
	if e.Err != nil {
		msg = fmt.Sprintf("%s error: %s: %v", e.Type, e.Message, e.Err)
	} else {
		msg = fmt.Sprintf("%s error: %s", e.Type, e.Message)
	}

	// Append context if available
	if len(e.Context) > 0 {
		msg += " ["
		first := true
		for k, v := range e.Context {
			if !first {
				msg += ", "
			}
			msg += fmt.Sprintf("%s=%v", k, v)
			first = false
		}
		msg += "]"
	}

	return msg
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Err
}

// WithContext adds context to the error
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// GetContext retrieves context from the error
func (e *Error) GetContext(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	val, ok := e.Context[key]
	return val, ok
}

// NewError creates a new domain error
func NewError(errType ErrorType, message string, err error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, err error) *Error {
	return NewError(ErrorTypeValidation, message, err)
}

// NewIOError creates an I/O error
func NewIOError(message string, err error) *Error {
	return NewError(ErrorTypeIO, message, err)
}

// NewProviderError creates a provider error
func NewProviderError(message string, err error) *Error {
	return NewError(ErrorTypeProvider, message, err)
}

// NewTranslationError creates a translation error
func NewTranslationError(message string, err error) *Error {
	return NewError(ErrorTypeTranslation, message, err)
}

// NewFormatError creates a format error
func NewFormatError(message string, err error) *Error {
	return NewError(ErrorTypeFormat, message, err)
}

// NewTerminologyError creates a terminology error
func NewTerminologyError(message string, err error) *Error {
	return NewError(ErrorTypeTerminology, message, err)
}

// NewConfigError creates a configuration error
func NewConfigError(message string, err error) *Error {
	return NewError(ErrorTypeConfig, message, err)
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errType ErrorType) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == errType
	}
	return false
}

// GetErrorType returns the error type if it's a domain error
func GetErrorType(err error) (ErrorType, bool) {
	if e, ok := err.(*Error); ok {
		return e.Type, true
	}
	return "", false
}

// WrapError wraps a standard error into a domain error
func WrapError(errType ErrorType, message string, err error) *Error {
	if err == nil {
		return nil
	}
	// If it's already a domain error, preserve the original type if not specified
	if domainErr, ok := err.(*Error); ok {
		if errType == "" {
			errType = domainErr.Type
		}
		return &Error{
			Type:    errType,
			Message: message,
			Err:     domainErr,
			Context: domainErr.Context,
		}
	}
	return NewError(errType, message, err)
}
