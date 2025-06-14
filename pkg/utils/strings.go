package utils

// DefaultString returns the value if it's not empty, otherwise returns the default value
func DefaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
