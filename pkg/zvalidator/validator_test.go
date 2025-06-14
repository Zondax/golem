package zvalidator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   *ValidationErrors
	}{
		{
			name:   "empty prefix",
			prefix: "",
			want: &ValidationErrors{
				errors: make([]ValidationError, 0),
				prefix: "",
			},
		},
		{
			name:   "with prefix",
			prefix: "test",
			want: &ValidationErrors{
				errors: make([]ValidationError, 0),
				prefix: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValidator(tt.prefix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidationErrors_AddError(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		message string
		want    []ValidationError
	}{
		{
			name:    "add single error",
			field:   "test_field",
			message: "is required",
			want: []ValidationError{
				{Field: "test_field", Message: "is required"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator("")
			result := v.AddError(tt.field, tt.message)
			assert.Equal(t, tt.want, v.errors)
			assert.Equal(t, v, result) // Test method chaining
		})
	}
}

func TestValidationErrors_AddErrorf(t *testing.T) {
	tests := []struct {
		name   string
		field  string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "formatted error with single argument",
			field:  "age",
			format: "must be at least %d years old",
			args:   []interface{}{18},
			want:   "must be at least 18 years old",
		},
		{
			name:   "formatted error with multiple arguments",
			field:  "score",
			format: "must be between %d and %d",
			args:   []interface{}{0, 100},
			want:   "must be between 0 and 100",
		},
		{
			name:   "formatted error with string argument",
			field:  "status",
			format: "invalid status '%s', expected 'active' or 'inactive'",
			args:   []interface{}{"pending"},
			want:   "invalid status 'pending', expected 'active' or 'inactive'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator("")
			result := v.AddErrorf(tt.field, tt.format, tt.args...)
			assert.Len(t, v.errors, 1)
			assert.Equal(t, tt.field, v.errors[0].Field)
			assert.Equal(t, tt.want, v.errors[0].Message)
			assert.Equal(t, v, result) // Test method chaining
		})
	}
}

func TestValidationErrors_Count(t *testing.T) {
	tests := []struct {
		name       string
		errorCount int
		wantCount  int
	}{
		{
			name:       "no errors",
			errorCount: 0,
			wantCount:  0,
		},
		{
			name:       "single error",
			errorCount: 1,
			wantCount:  1,
		},
		{
			name:       "multiple errors",
			errorCount: 5,
			wantCount:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator("")
			for i := 0; i < tt.errorCount; i++ {
				v.AddErrorf("field%d", "error %d", i)
			}
			assert.Equal(t, tt.wantCount, v.Count())
		})
	}
}

func TestValidationErrors_ValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{
			name:      "empty string should add error",
			field:     "name",
			value:     "",
			wantError: true,
		},
		{
			name:      "whitespace only should add error",
			field:     "name",
			value:     "   ",
			wantError: true,
		},
		{
			name:      "tabs and newlines should add error",
			field:     "name",
			value:     "\t\n  \r",
			wantError: true,
		},
		{
			name:      "valid string should not add error",
			field:     "name",
			value:     "John Doe",
			wantError: false,
		},
		{
			name:      "string with leading/trailing spaces but content should not add error",
			field:     "name",
			value:     "  John  ",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator("")
			result := v.ValidateRequired(tt.field, tt.value)

			if tt.wantError {
				assert.True(t, v.HasErrors())
				assert.Equal(t, 1, v.Count())
				assert.Equal(t, tt.field, v.errors[0].Field)
				assert.Equal(t, "is required", v.errors[0].Message)
			} else {
				assert.False(t, v.HasErrors())
				assert.Equal(t, 0, v.Count())
			}
			assert.Equal(t, v, result) // Test method chaining
		})
	}
}

func TestValidationErrors_ValidateMinLength(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		minLength int
		wantError bool
		wantMsg   string
	}{
		{
			name:      "string too short should add error",
			field:     "password",
			value:     "123",
			minLength: 8,
			wantError: true,
			wantMsg:   "must be at least 8 characters",
		},
		{
			name:      "empty string with min length should add error",
			field:     "password",
			value:     "",
			minLength: 1,
			wantError: true,
			wantMsg:   "must be at least 1 characters",
		},
		{
			name:      "string exactly min length should not add error",
			field:     "password",
			value:     "12345678",
			minLength: 8,
			wantError: false,
		},
		{
			name:      "string longer than min length should not add error",
			field:     "password",
			value:     "123456789",
			minLength: 8,
			wantError: false,
		},
		{
			name:      "zero min length should never add error",
			field:     "optional",
			value:     "",
			minLength: 0,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator("")
			result := v.ValidateMinLength(tt.field, tt.value, tt.minLength)

			if tt.wantError {
				assert.True(t, v.HasErrors())
				assert.Equal(t, 1, v.Count())
				assert.Equal(t, tt.field, v.errors[0].Field)
				assert.Equal(t, tt.wantMsg, v.errors[0].Message)
			} else {
				assert.False(t, v.HasErrors())
				assert.Equal(t, 0, v.Count())
			}
			assert.Equal(t, v, result) // Test method chaining
		})
	}
}

func TestValidationErrors_ValidateMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		maxLength int
		wantError bool
		wantMsg   string
	}{
		{
			name:      "string too long should add error",
			field:     "bio",
			value:     "This is a very long biography that exceeds the maximum allowed length",
			maxLength: 50,
			wantError: true,
			wantMsg:   "must not exceed 50 characters",
		},
		{
			name:      "string exactly max length should not add error",
			field:     "bio",
			value:     "This bio is exactly fifty characters long here.",
			maxLength: 47,
			wantError: false,
		},
		{
			name:      "string shorter than max length should not add error",
			field:     "bio",
			value:     "Short bio",
			maxLength: 50,
			wantError: false,
		},
		{
			name:      "empty string should not add error",
			field:     "bio",
			value:     "",
			maxLength: 10,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator("")
			result := v.ValidateMaxLength(tt.field, tt.value, tt.maxLength)

			if tt.wantError {
				assert.True(t, v.HasErrors())
				assert.Equal(t, 1, v.Count())
				assert.Equal(t, tt.field, v.errors[0].Field)
				assert.Equal(t, tt.wantMsg, v.errors[0].Message)
			} else {
				assert.False(t, v.HasErrors())
				assert.Equal(t, 0, v.Count())
			}
			assert.Equal(t, v, result) // Test method chaining
		})
	}
}

func TestValidationErrors_MethodChaining(t *testing.T) {
	t.Run("fluent API with method chaining", func(t *testing.T) {
		v := NewValidator("user").
			AddError("field1", "error1").
			AddErrorf("field2", "error %d", 2).
			ValidateRequired("name", "").
			ValidateMinLength("password", "123", 8).
			ValidateMaxLength("bio", "This is a very long bio that definitely exceeds the fifty character limit", 50)

		assert.True(t, v.HasErrors())
		assert.Equal(t, 5, v.Count())

		// Verify all errors were added correctly
		expectedErrors := []struct{ field, message string }{
			{"field1", "error1"},
			{"field2", "error 2"},
			{"name", "is required"},
			{"password", "must be at least 8 characters"},
			{"bio", "must not exceed 50 characters"},
		}

		for i, expected := range expectedErrors {
			assert.Equal(t, expected.field, v.errors[i].Field)
			assert.Equal(t, expected.message, v.errors[i].Message)
		}
	})

	t.Run("chaining with no errors when validation passes", func(t *testing.T) {
		v := NewValidator("product").
			ValidateRequired("name", "Product Name").
			ValidateMinLength("description", "This is a long enough description", 10).
			ValidateMaxLength("title", "Short", 50)

		assert.False(t, v.HasErrors())
		assert.Equal(t, 0, v.Count())
		assert.Nil(t, v.Error())
	})
}

func TestValidationErrors_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		addError bool
		want     bool
	}{
		{
			name:     "no errors",
			addError: false,
			want:     false,
		},
		{
			name:     "has errors",
			addError: true,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator("")
			if tt.addError {
				v.AddError("test", "error")
			}
			assert.Equal(t, tt.want, v.HasErrors())
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		errors   []struct{ field, message string }
		wantErr  bool
		wantText string
	}{
		{
			name:     "no errors",
			prefix:   "",
			errors:   nil,
			wantErr:  false,
			wantText: "",
		},
		{
			name:   "single error without prefix",
			prefix: "",
			errors: []struct{ field, message string }{
				{field: "test", message: "is required"},
			},
			wantErr:  true,
			wantText: "validation failed: test: is required",
		},
		{
			name:   "multiple errors with prefix",
			prefix: "user",
			errors: []struct{ field, message string }{
				{field: "name", message: "is required"},
				{field: "email", message: "invalid format"},
			},
			wantErr:  true,
			wantText: "user validation failed: name: is required, email: invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(tt.prefix)
			for _, e := range tt.errors {
				v.AddError(e.field, e.message)
			}

			err := v.Error()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantText, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationErrors_Integration(t *testing.T) {
	t.Run("real-world user validation scenario", func(t *testing.T) {
		// Simulating a real-world validation scenario
		v := NewValidator("user")

		// Add multiple validation errors
		v.AddError("name", "is required")
		v.AddError("age", "must be greater than 0")
		v.AddError("email", "invalid format")

		// Check if has errors
		assert.True(t, v.HasErrors())
		assert.Equal(t, 3, v.Count())

		// Verify error message format
		expectedErr := "user validation failed: name: is required, age: must be greater than 0, email: invalid format"
		assert.Equal(t, expectedErr, v.Error().Error())

		// Create a new validator without errors
		v2 := NewValidator("test")
		assert.False(t, v2.HasErrors())
		assert.Equal(t, 0, v2.Count())
		assert.Nil(t, v2.Error())
	})

	t.Run("comprehensive validation with convenience methods", func(t *testing.T) {
		// Test realistic user registration validation
		userData := struct {
			name     string
			email    string
			password string
			bio      string
			age      int
		}{
			name:     "",
			email:    "invalid-email",
			password: "123",
			bio:      "This is a very long biography that exceeds the maximum allowed length for user profiles in our system",
			age:      -5,
		}

		v := NewValidator("registration").
			ValidateRequired("name", userData.name).
			ValidateMinLength("password", userData.password, 8).
			ValidateMaxLength("bio", userData.bio, 50).
			AddErrorf("age", "must be between %d and %d", 0, 120)

		assert.True(t, v.HasErrors())
		assert.Equal(t, 4, v.Count())

		err := v.Error()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "registration validation failed")
		assert.Contains(t, err.Error(), "name: is required")
		assert.Contains(t, err.Error(), "password: must be at least 8 characters")
		assert.Contains(t, err.Error(), "bio: must not exceed 50 characters")
		assert.Contains(t, err.Error(), "age: must be between 0 and 120")
	})

	t.Run("successful validation scenario", func(t *testing.T) {
		// Test successful validation with no errors
		userData := struct {
			name     string
			password string
			bio      string
		}{
			name:     "John Doe",
			password: "securepassword123",
			bio:      "Software developer",
		}

		v := NewValidator("user").
			ValidateRequired("name", userData.name).
			ValidateMinLength("password", userData.password, 8).
			ValidateMaxLength("bio", userData.bio, 100)

		assert.False(t, v.HasErrors())
		assert.Equal(t, 0, v.Count())
		assert.Nil(t, v.Error())
	})
}
