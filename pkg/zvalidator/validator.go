package zvalidator

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors struct {
	errors []ValidationError
	prefix string
}

// NewValidator creates a new validator with an optional prefix for error messages
func NewValidator(prefix string) *ValidationErrors {
	return &ValidationErrors{
		errors: make([]ValidationError, 0),
		prefix: prefix,
	}
}

// AddError adds a validation error and returns the validator for method chaining
func (v *ValidationErrors) AddError(field, message string) *ValidationErrors {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
	return v
}

// AddErrorf adds a formatted validation error and returns the validator for method chaining
func (v *ValidationErrors) AddErrorf(field, format string, args ...interface{}) *ValidationErrors {
	return v.AddError(field, fmt.Sprintf(format, args...))
}

// HasErrors returns true if there are validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.errors) > 0
}

// Count returns the number of validation errors
func (v *ValidationErrors) Count() int {
	return len(v.errors)
}

// Error returns all validation errors as a single error message
func (v *ValidationErrors) Error() error {
	if !v.HasErrors() {
		return nil
	}

	var messages []string
	for _, err := range v.errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}

	prefix := "validation failed"
	if v.prefix != "" {
		prefix = fmt.Sprintf("%s validation failed", v.prefix)
	}

	return fmt.Errorf("%s: %s", prefix, strings.Join(messages, ", "))
}

// ValidateRequired is a convenience method for required field validation
func (v *ValidationErrors) ValidateRequired(field, value string) *ValidationErrors {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
	}
	return v
}

// ValidateMinLength is a convenience method for minimum length validation
func (v *ValidationErrors) ValidateMinLength(field, value string, minLength int) *ValidationErrors {
	if len(value) < minLength {
		v.AddErrorf(field, "must be at least %d characters", minLength)
	}
	return v
}

// ValidateMaxLength is a convenience method for maximum length validation
func (v *ValidationErrors) ValidateMaxLength(field, value string, maxLength int) *ValidationErrors {
	if len(value) > maxLength {
		v.AddErrorf(field, "must not exceed %d characters", maxLength)
	}
	return v
}
