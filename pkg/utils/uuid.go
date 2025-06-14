// Package utils provides utility functions for common operations including UUID generation.
package utils

import (
	"fmt"

	"github.com/google/uuid"
)

// GenerateUUID generates a random UUID v4 string.
// Returns a 36-character UUID in format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx".
//
// Example:
//
//	id := utils.GenerateUUID()
//	// Output: "550e8400-e29b-41d4-a716-446655440000"
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateUUIDShort generates a short 8-character UUID string.
// Higher collision probability than full UUID - use carefully in production.
//
// Example:
//
//	id := utils.GenerateUUIDShort()
//	// Output: "550e8400"
func GenerateUUIDShort() string {
	return uuid.New().String()[:8]
}

// GenerateUUIDv5 generates a deterministic UUID v5 from namespace and fields.
// Same inputs always produce the same UUID. Fields are joined with '|' separator.
//
// Parameters:
//   - namespace: UUID namespace (e.g., uuid.NameSpaceDNS)
//   - fields: Non-empty strings to generate UUID from
//
// Returns error if fields is empty or contains empty strings.
//
// Example:
//
//	id, err := utils.GenerateUUIDv5(uuid.NameSpaceDNS, []string{"user", "john.doe"})
func GenerateUUIDv5(namespace uuid.UUID, fields []string) (string, error) {
	if len(fields) == 0 {
		return "", fmt.Errorf("at least one field is required for UUID generation")
	}

	for i, field := range fields {
		if field == "" {
			return "", fmt.Errorf("field at position %d cannot be empty", i)
		}
	}

	// Join fields with separator
	var payload string
	for i, field := range fields {
		if i == 0 {
			payload = field
		} else {
			payload = fmt.Sprintf("%s|%s", payload, field)
		}
	}

	// Generate UUID v5
	u := uuid.NewSHA1(namespace, []byte(payload))
	return u.String(), nil
}
