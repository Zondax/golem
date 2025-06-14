package utils

import (
	"fmt"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateUUIDShort() string {
	return uuid.New().String()[:8]
}

// GenerateUUIDv5 generates a deterministic UUID v5 based on the provided namespace and fields.
// The fields are joined with a '|' separator before generating the UUID.
// Returns an error if any field is empty.
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
