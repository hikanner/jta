package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		contains []string
	}{
		{
			name: "error with message only",
			err: &Error{
				Type:    ErrorTypeValidation,
				Message: "invalid input",
			},
			contains: []string{"validation error", "invalid input"},
		},
		{
			name: "error with wrapped error",
			err: &Error{
				Type:    ErrorTypeIO,
				Message: "file not found",
				Err:     errors.New("open failed"),
			},
			contains: []string{"io error", "file not found", "open failed"},
		},
		{
			name: "error with context",
			err: &Error{
				Type:    ErrorTypeProvider,
				Message: "API failed",
				Context: map[string]any{
					"provider": "openai",
					"status":   500,
				},
			},
			contains: []string{"provider error", "API failed", "provider=openai", "status=500"},
		},
		{
			name: "error with everything",
			err: &Error{
				Type:    ErrorTypeTranslation,
				Message: "translation failed",
				Err:     errors.New("timeout"),
				Context: map[string]any{
					"lang": "en",
				},
			},
			contains: []string{"translation error", "translation failed", "timeout", "lang=en"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Error() = %q, should contain %q", result, expected)
				}
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := &Error{
		Type:    ErrorTypeValidation,
		Message: "wrapped",
		Err:     originalErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test nil unwrap
	err2 := &Error{
		Type:    ErrorTypeValidation,
		Message: "no wrap",
	}
	if err2.Unwrap() != nil {
		t.Errorf("Unwrap() should return nil for non-wrapped error")
	}
}

func TestError_WithContext(t *testing.T) {
	err := &Error{
		Type:    ErrorTypeValidation,
		Message: "test",
	}

	// Add context
	_ = err.WithContext("key1", "value1")
	_ = err.WithContext("key2", 123)

	// Verify context
	if val, ok := err.GetContext("key1"); !ok || val != "value1" {
		t.Errorf("GetContext(key1) = %v, %v, want value1, true", val, ok)
	}
	if val, ok := err.GetContext("key2"); !ok || val != 123 {
		t.Errorf("GetContext(key2) = %v, %v, want 123, true", val, ok)
	}

	// Test non-existent key
	if val, ok := err.GetContext("nonexistent"); ok {
		t.Errorf("GetContext(nonexistent) = %v, true, want _, false", val)
	}
}

func TestError_GetContext(t *testing.T) {
	tests := []struct {
		name      string
		err       *Error
		key       string
		wantValue any
		wantOk    bool
	}{
		{
			name: "existing key",
			err: &Error{
				Type:    ErrorTypeValidation,
				Message: "test",
				Context: map[string]any{"key": "value"},
			},
			key:       "key",
			wantValue: "value",
			wantOk:    true,
		},
		{
			name: "non-existent key",
			err: &Error{
				Type:    ErrorTypeValidation,
				Message: "test",
				Context: map[string]any{"key": "value"},
			},
			key:       "missing",
			wantValue: nil,
			wantOk:    false,
		},
		{
			name: "nil context",
			err: &Error{
				Type:    ErrorTypeValidation,
				Message: "test",
			},
			key:       "any",
			wantValue: nil,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := tt.err.GetContext(tt.key)
			if gotOk != tt.wantOk {
				t.Errorf("GetContext() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotValue != tt.wantValue {
				t.Errorf("GetContext() value = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	originalErr := errors.New("original")
	err := NewError(ErrorTypeValidation, "test message", originalErr)

	if err.Type != ErrorTypeValidation {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeValidation)
	}
	if err.Message != "test message" {
		t.Errorf("Message = %v, want %v", err.Message, "test message")
	}
	if err.Err != originalErr {
		t.Errorf("Err = %v, want %v", err.Err, originalErr)
	}
	if err.Context == nil {
		t.Error("Context should be initialized")
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("invalid", nil)
	if err.Type != ErrorTypeValidation {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeValidation)
	}
}

func TestNewIOError(t *testing.T) {
	err := NewIOError("file error", nil)
	if err.Type != ErrorTypeIO {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeIO)
	}
}

func TestNewProviderError(t *testing.T) {
	err := NewProviderError("api error", nil)
	if err.Type != ErrorTypeProvider {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeProvider)
	}
}

func TestNewTranslationError(t *testing.T) {
	err := NewTranslationError("translation failed", nil)
	if err.Type != ErrorTypeTranslation {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeTranslation)
	}
}

func TestNewFormatError(t *testing.T) {
	err := NewFormatError("format issue", nil)
	if err.Type != ErrorTypeFormat {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeFormat)
	}
}

func TestNewTerminologyError(t *testing.T) {
	err := NewTerminologyError("term error", nil)
	if err.Type != ErrorTypeTerminology {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeTerminology)
	}
}

func TestNewConfigError(t *testing.T) {
	err := NewConfigError("config error", nil)
	if err.Type != ErrorTypeConfig {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeConfig)
	}
}

func TestIsErrorType(t *testing.T) {
	validationErr := NewValidationError("test", nil)
	ioErr := NewIOError("test", nil)
	standardErr := errors.New("standard error")

	tests := []struct {
		name     string
		err      error
		errType  ErrorType
		expected bool
	}{
		{"matching type", validationErr, ErrorTypeValidation, true},
		{"non-matching type", validationErr, ErrorTypeIO, false},
		{"standard error", standardErr, ErrorTypeValidation, false},
		{"nil error", nil, ErrorTypeValidation, false},
		{"different domain error", ioErr, ErrorTypeIO, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsErrorType(tt.err, tt.errType)
			if result != tt.expected {
				t.Errorf("IsErrorType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedType ErrorType
		expectedOk   bool
	}{
		{
			name:         "validation error",
			err:          NewValidationError("test", nil),
			expectedType: ErrorTypeValidation,
			expectedOk:   true,
		},
		{
			name:         "io error",
			err:          NewIOError("test", nil),
			expectedType: ErrorTypeIO,
			expectedOk:   true,
		},
		{
			name:         "standard error",
			err:          errors.New("standard"),
			expectedType: "",
			expectedOk:   false,
		},
		{
			name:         "nil error",
			err:          nil,
			expectedType: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotOk := GetErrorType(tt.err)
			if gotOk != tt.expectedOk {
				t.Errorf("GetErrorType() ok = %v, want %v", gotOk, tt.expectedOk)
			}
			if gotType != tt.expectedType {
				t.Errorf("GetErrorType() type = %v, want %v", gotType, tt.expectedType)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name         string
		errType      ErrorType
		message      string
		originalErr  error
		expectedType ErrorType
		expectNil    bool
	}{
		{
			name:         "wrap standard error",
			errType:      ErrorTypeValidation,
			message:      "wrapped",
			originalErr:  errors.New("original"),
			expectedType: ErrorTypeValidation,
			expectNil:    false,
		},
		{
			name:         "wrap domain error with type",
			errType:      ErrorTypeIO,
			message:      "wrapped",
			originalErr:  NewValidationError("original", nil),
			expectedType: ErrorTypeIO,
			expectNil:    false,
		},
		{
			name:         "wrap domain error without type",
			errType:      "",
			message:      "wrapped",
			originalErr:  NewValidationError("original", nil),
			expectedType: ErrorTypeValidation,
			expectNil:    false,
		},
		{
			name:        "wrap nil error",
			errType:     ErrorTypeValidation,
			message:     "wrapped",
			originalErr: nil,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.errType, tt.message, tt.originalErr)

			if tt.expectNil {
				if result != nil {
					t.Errorf("WrapError() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("WrapError() returned nil, want error")
			}

			if result.Type != tt.expectedType {
				t.Errorf("WrapError().Type = %v, want %v", result.Type, tt.expectedType)
			}

			if result.Message != tt.message {
				t.Errorf("WrapError().Message = %v, want %v", result.Message, tt.message)
			}

			// Check that context is preserved from domain errors
			if domainErr, ok := tt.originalErr.(*Error); ok {
				if result.Context == nil && domainErr.Context != nil {
					t.Error("WrapError() should preserve context from domain error")
				}
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Ensure all error type constants are defined
	types := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeIO,
		ErrorTypeProvider,
		ErrorTypeTranslation,
		ErrorTypeFormat,
		ErrorTypeTerminology,
		ErrorTypeConfig,
	}

	expectedValues := []string{
		"validation",
		"io",
		"provider",
		"translation",
		"format",
		"terminology",
		"config",
	}

	if len(types) != len(expectedValues) {
		t.Fatalf("Expected %d error types, got %d", len(expectedValues), len(types))
	}

	for i, errType := range types {
		if string(errType) != expectedValues[i] {
			t.Errorf("ErrorType[%d] = %q, want %q", i, errType, expectedValues[i])
		}
	}
}
