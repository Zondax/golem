package zconverters

import (
	"errors"
	"math"
)

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

// Int64ToUint64 converts int64 to uint64 safely, returning 0 if the value is negative
func Int64ToUint64(value int64) uint64 {
	if value >= 0 {
		return uint64(value)
	}
	return 0
}

// IntToInt32 converts int to int32 safely, capping at MaxInt32 if the value is too large
func IntToInt32(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	return int32(value)
}

// LenToInt32 converts the length of a slice/array to int32 safely, capping at MaxInt32
func LenToInt32(length int) int32 {
	return IntToInt32(length)
}
