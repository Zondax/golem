package zvalidator

import (
	"fmt"
	"net/mail"
	"strings"
)

const (
	maxEmailLength = 254 // RFC 5321
)

// ValidateEmail performs email validation using Go's standard library
// with additional business logic constraints
func ValidateEmail(email string) error {
	// Basic empty check
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Length check (business constraint)
	if len(email) > maxEmailLength {
		return fmt.Errorf("email exceeds maximum length of %d characters", maxEmailLength)
	}

	// Use Go's standard email validation (RFC 5322 compliant)
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}

	// Additional business logic: ensure the parsed address matches input
	// (ParseAddress can be lenient and "fix" some formats)
	if addr.Address != email {
		return fmt.Errorf("email format was modified during parsing, original: %s, parsed: %s", email, addr.Address)
	}

	// Additional business constraint: require domain with dot
	if !strings.Contains(email, ".") {
		return fmt.Errorf("email domain must contain at least one dot")
	}

	return nil
}
