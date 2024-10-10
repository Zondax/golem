package utils

import "errors"

// IntToUInt64 safely converts an int to uint64.
// Returns an error if the number is negative.
func IntToUInt64(i int) (uint64, error) {
	if i < 0 {
		return 0, errors.New("cannot convert negative int to uint64")
	}
	return uint64(i), nil
}

// IntToUInt safely converts an int to uint.
// Returns an error if the number is negative.
func IntToUInt(i int) (uint, error) {
	if i < 0 {
		return 0, errors.New("cannot convert negative int to uint")
	}
	return uint(i), nil
}
