// Package zptr provides utility functions for working with pointers.
package zptr

// StringToPtr takes a string value and returns a pointer to a copy of it.
func StringToPtr(s string) *string {
	return &s
}

// StringOrDefault returns the value pointed to by a string pointer, or an empty string if the pointer is nil.
func StringOrDefault(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BoolToPtr takes a boolean value and returns a pointer to a copy of it.
func BoolToPtr(b bool) *bool {
	return &b
}

// BoolOrDefault returns the value pointed to by a bool pointer, or 'false' if the pointer is nil.
func BoolOrDefault(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// IntToPtr takes an integer value and returns a pointer to a copy of it.
func IntToPtr(i int) *int {
	return &i
}

// IntOrDefault returns the value pointed to by an int pointer, or 0 if the pointer is nil.
func IntOrDefault(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// Float32ToPtr takes a float32 value and returns a pointer to a copy of it.
func Float32ToPtr(f float32) *float32 {
	return &f
}

// Float32OrDefault returns the value pointed to by a float32 pointer, or 0.0 if the pointer is nil.
func Float32OrDefault(f *float32) float32 {
	if f == nil {
		return 0.0
	}
	return *f
}

// Float64ToPtr takes a float64 value and returns a pointer to a copy of it.
func Float64ToPtr(f float64) *float64 {
	return &f
}

// Float64OrDefault returns the value pointed to by a float64 pointer, or 0.0 if the pointer is nil.
func Float64OrDefault(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}
