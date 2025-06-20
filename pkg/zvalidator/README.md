# zvalidator

A simple, idiomatic Go validation package that provides fluent API for collecting and managing validation errors.

## Features

- **Fluent API**: Method chaining for readable validation code
- **Error Collection**: Accumulate multiple validation errors
- **Convenience Methods**: Common validation patterns built-in
- **Email Validation**: RFC-compliant email validation with business constraints
- **Formatted Errors**: Support for formatted error messages
- **Zero Dependencies**: Uses only Go standard library (except for tests)

## Installation

```bash
go get github.com/zondax/golem/pkg/zvalidator
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/zondax/golem/pkg/zvalidator"
)

func main() {
    validator := zvalidator.NewValidator("user")
    validator.AddError("name", "is required")
    validator.AddError("email", "invalid format")
    
    if validator.HasErrors() {
        fmt.Println(validator.Error())
        // Output: user validation failed: name: is required, email: invalid format
    }
}
```

### Fluent API with Method Chaining

```go
func validateUser(name, email, password, bio string, age int) error {
    validator := zvalidator.NewValidator("user").
        ValidateRequired("name", name).
        ValidateRequired("email", email).
        ValidateMinLength("password", password, 8).
        ValidateMaxLength("bio", bio, 500).
        AddErrorf("age", "must be between %d and %d", 18, 120)
    
    if validator.HasErrors() {
        return validator.Error()
    }
    return nil
}
```

### Email Validation

```go
func validateEmail(email string) error {
    return zvalidator.ValidateEmail(email)
}

// Examples:
// validateEmail("user@example.com") // nil
// validateEmail("invalid-email")    // error: invalid email format
// validateEmail("")                 // error: email is required
```

## API Reference

### Core Types

#### ValidationError
```go
type ValidationError struct {
    Field   string
    Message string
}
```

#### ValidationErrors
```go
type ValidationErrors struct {
    // private fields
}
```

### Constructor

#### NewValidator
```go
func NewValidator(prefix string) *ValidationErrors
```
Creates a new validator with an optional prefix for error messages.

### Methods

#### AddError
```go
func (v *ValidationErrors) AddError(field, message string) *ValidationErrors
```
Adds a validation error and returns the validator for method chaining.

#### AddErrorf
```go
func (v *ValidationErrors) AddErrorf(field, format string, args ...interface{}) *ValidationErrors
```
Adds a formatted validation error and returns the validator for method chaining.

#### HasErrors
```go
func (v *ValidationErrors) HasErrors() bool
```
Returns true if there are validation errors.

#### Count
```go
func (v *ValidationErrors) Count() int
```
Returns the number of validation errors.

#### Error
```go
func (v *ValidationErrors) Error() error
```
Returns all validation errors as a single error message, or nil if no errors.

### Convenience Methods

#### ValidateRequired
```go
func (v *ValidationErrors) ValidateRequired(field, value string) *ValidationErrors
```
Validates that a field is not empty (after trimming whitespace).

#### ValidateMinLength
```go
func (v *ValidationErrors) ValidateMinLength(field, value string, minLength int) *ValidationErrors
```
Validates that a field meets minimum length requirements.

#### ValidateMaxLength
```go
func (v *ValidationErrors) ValidateMaxLength(field, value string, maxLength int) *ValidationErrors
```
Validates that a field doesn't exceed maximum length.

### Standalone Functions

#### ValidateEmail
```go
func ValidateEmail(email string) error
```
Performs RFC 5322 compliant email validation with additional business constraints:
- Maximum length of 254 characters (RFC 5321)
- Requires domain with at least one dot
- Strict format checking

## Examples

### User Registration Validation

```go
func validateUserRegistration(req UserRegistrationRequest) error {
    validator := zvalidator.NewValidator("registration").
        ValidateRequired("username", req.Username).
        ValidateMinLength("username", req.Username, 3).
        ValidateMaxLength("username", req.Username, 50).
        ValidateRequired("email", req.Email).
        ValidateMinLength("password", req.Password, 8).
        ValidateRequired("firstName", req.FirstName).
        ValidateRequired("lastName", req.LastName)
    
    // Custom email validation
    if req.Email != "" {
        if err := zvalidator.ValidateEmail(req.Email); err != nil {
            validator.AddError("email", err.Error())
        }
    }
    
    // Custom age validation
    if req.Age < 18 || req.Age > 120 {
        validator.AddErrorf("age", "must be between %d and %d", 18, 120)
    }
    
    if validator.HasErrors() {
        return validator.Error()
    }
    
    return nil
}
```

### API Response Validation

```go
func validateAPIResponse(data map[string]interface{}) error {
    validator := zvalidator.NewValidator("API response")
    
    // Check required fields
    requiredFields := []string{"id", "name", "status"}
    for _, field := range requiredFields {
        if _, exists := data[field]; !exists {
            validator.AddErrorf(field, "is required in response")
        }
    }
    
    // Validate specific field types
    if id, ok := data["id"].(string); ok {
        validator.ValidateRequired("id", id)
    } else {
        validator.AddError("id", "must be a string")
    }
    
    return validator.Error()
}
```

## Design Principles

- **Simplicity First**: Focused on essential validation features without over-engineering
- **Method Chaining**: Fluent interface for better code readability
- **Go Idioms**: Follows standard Go patterns and conventions
- **Zero Breaking Changes**: Backward compatible API design
- **Performance**: Minimal allocations and efficient error collection

## Testing

The package includes comprehensive tests with 100% code coverage:

```bash
go test ./pkg/zvalidator/
go test -cover ./pkg/zvalidator/
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This package is part of the Zondax Golem project and follows the same licensing terms.
