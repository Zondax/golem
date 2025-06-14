package zconverters

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntToUInt64(t *testing.T) {
	t.Run("ShouldConvertPositiveInt", func(t *testing.T) {
		result, err := IntToUInt64(42)
		
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), result)
	})
	
	t.Run("ShouldConvertZero", func(t *testing.T) {
		result, err := IntToUInt64(0)
		
		assert.NoError(t, err)
		assert.Equal(t, uint64(0), result)
	})
	
	t.Run("ShouldConvertMaxInt", func(t *testing.T) {
		result, err := IntToUInt64(math.MaxInt)
		
		assert.NoError(t, err)
		assert.Equal(t, uint64(math.MaxInt), result)
	})
	
	t.Run("ShouldReturnErrorForNegativeInt", func(t *testing.T) {
		result, err := IntToUInt64(-1)
		
		assert.Error(t, err)
		assert.Equal(t, uint64(0), result)
		assert.Contains(t, err.Error(), "cannot convert negative int to uint64")
	})
	
	t.Run("ShouldReturnErrorForMinInt", func(t *testing.T) {
		result, err := IntToUInt64(math.MinInt)
		
		assert.Error(t, err)
		assert.Equal(t, uint64(0), result)
		assert.Contains(t, err.Error(), "cannot convert negative int to uint64")
	})
	
	t.Run("ShouldReturnErrorForLargeNegativeInt", func(t *testing.T) {
		result, err := IntToUInt64(-999999)
		
		assert.Error(t, err)
		assert.Equal(t, uint64(0), result)
		assert.Contains(t, err.Error(), "cannot convert negative int to uint64")
	})
}

func TestIntToUInt(t *testing.T) {
	t.Run("ShouldConvertPositiveInt", func(t *testing.T) {
		result, err := IntToUInt(42)
		
		assert.NoError(t, err)
		assert.Equal(t, uint(42), result)
	})
	
	t.Run("ShouldConvertZero", func(t *testing.T) {
		result, err := IntToUInt(0)
		
		assert.NoError(t, err)
		assert.Equal(t, uint(0), result)
	})
	
	t.Run("ShouldConvertMaxInt", func(t *testing.T) {
		result, err := IntToUInt(math.MaxInt)
		
		assert.NoError(t, err)
		assert.Equal(t, uint(math.MaxInt), result)
	})
	
	t.Run("ShouldReturnErrorForNegativeInt", func(t *testing.T) {
		result, err := IntToUInt(-1)
		
		assert.Error(t, err)
		assert.Equal(t, uint(0), result)
		assert.Contains(t, err.Error(), "cannot convert negative int to uint")
	})
	
	t.Run("ShouldReturnErrorForMinInt", func(t *testing.T) {
		result, err := IntToUInt(math.MinInt)
		
		assert.Error(t, err)
		assert.Equal(t, uint(0), result)
		assert.Contains(t, err.Error(), "cannot convert negative int to uint")
	})
	
	t.Run("ShouldReturnErrorForLargeNegativeInt", func(t *testing.T) {
		result, err := IntToUInt(-123456)
		
		assert.Error(t, err)
		assert.Equal(t, uint(0), result)
		assert.Contains(t, err.Error(), "cannot convert negative int to uint")
	})
}

func TestInt64ToUint64(t *testing.T) {
	t.Run("ShouldConvertPositiveInt64", func(t *testing.T) {
		result := Int64ToUint64(42)
		
		assert.Equal(t, uint64(42), result)
	})
	
	t.Run("ShouldConvertZero", func(t *testing.T) {
		result := Int64ToUint64(0)
		
		assert.Equal(t, uint64(0), result)
	})
	
	t.Run("ShouldConvertMaxInt64", func(t *testing.T) {
		result := Int64ToUint64(math.MaxInt64)
		
		assert.Equal(t, uint64(math.MaxInt64), result)
	})
	
	t.Run("ShouldReturnZeroForNegativeInt64", func(t *testing.T) {
		result := Int64ToUint64(-1)
		
		assert.Equal(t, uint64(0), result)
	})
	
	t.Run("ShouldReturnZeroForMinInt64", func(t *testing.T) {
		result := Int64ToUint64(math.MinInt64)
		
		assert.Equal(t, uint64(0), result)
	})
	
	t.Run("ShouldReturnZeroForLargeNegativeInt64", func(t *testing.T) {
		result := Int64ToUint64(-999999999)
		
		assert.Equal(t, uint64(0), result)
	})
	
	t.Run("ShouldHandleLargePositiveValues", func(t *testing.T) {
		largeValue := int64(1<<62 - 1) // Large positive value
		result := Int64ToUint64(largeValue)
		
		assert.Equal(t, uint64(largeValue), result)
	})
}

func TestIntToInt32(t *testing.T) {
	t.Run("ShouldConvertNormalInt", func(t *testing.T) {
		result := IntToInt32(42)
		
		assert.Equal(t, int32(42), result)
	})
	
	t.Run("ShouldConvertZero", func(t *testing.T) {
		result := IntToInt32(0)
		
		assert.Equal(t, int32(0), result)
	})
	
	t.Run("ShouldConvertNegativeInt", func(t *testing.T) {
		result := IntToInt32(-42)
		
		assert.Equal(t, int32(-42), result)
	})
	
	t.Run("ShouldConvertMaxInt32", func(t *testing.T) {
		result := IntToInt32(math.MaxInt32)
		
		assert.Equal(t, int32(math.MaxInt32), result)
	})
	
	t.Run("ShouldConvertMinInt32", func(t *testing.T) {
		result := IntToInt32(math.MinInt32)
		
		assert.Equal(t, int32(math.MinInt32), result)
	})
	
	t.Run("ShouldCapAtMaxInt32WhenTooLarge", func(t *testing.T) {
		result := IntToInt32(math.MaxInt32 + 1)
		
		assert.Equal(t, int32(math.MaxInt32), result)
	})
	
	t.Run("ShouldCapAtMinInt32WhenTooSmall", func(t *testing.T) {
		result := IntToInt32(math.MinInt32 - 1)
		
		assert.Equal(t, int32(math.MinInt32), result)
	})
	
	t.Run("ShouldCapVeryLargePositiveValues", func(t *testing.T) {
		result := IntToInt32(math.MaxInt)
		
		assert.Equal(t, int32(math.MaxInt32), result)
	})
	
	t.Run("ShouldCapVeryLargeNegativeValues", func(t *testing.T) {
		result := IntToInt32(math.MinInt)
		
		assert.Equal(t, int32(math.MinInt32), result)
	})
	
	t.Run("ShouldHandleValuesJustWithinRange", func(t *testing.T) {
		// Test values just within the int32 range
		result1 := IntToInt32(math.MaxInt32 - 1)
		result2 := IntToInt32(math.MinInt32 + 1)
		
		assert.Equal(t, int32(math.MaxInt32-1), result1)
		assert.Equal(t, int32(math.MinInt32+1), result2)
	})
}

func TestLenToInt32(t *testing.T) {
	t.Run("ShouldConvertSmallLength", func(t *testing.T) {
		result := LenToInt32(10)
		
		assert.Equal(t, int32(10), result)
	})
	
	t.Run("ShouldConvertZeroLength", func(t *testing.T) {
		result := LenToInt32(0)
		
		assert.Equal(t, int32(0), result)
	})
	
	t.Run("ShouldConvertMaxInt32Length", func(t *testing.T) {
		result := LenToInt32(math.MaxInt32)
		
		assert.Equal(t, int32(math.MaxInt32), result)
	})
	
	t.Run("ShouldCapAtMaxInt32WhenTooLarge", func(t *testing.T) {
		result := LenToInt32(math.MaxInt32 + 1000)
		
		assert.Equal(t, int32(math.MaxInt32), result)
	})
	
	t.Run("ShouldCapVeryLargeLength", func(t *testing.T) {
		result := LenToInt32(math.MaxInt)
		
		assert.Equal(t, int32(math.MaxInt32), result)
	})
	
	t.Run("ShouldHandleNegativeLength", func(t *testing.T) {
		// Although negative lengths don't make practical sense,
		// the function should handle them gracefully
		result := LenToInt32(-1)
		
		assert.Equal(t, int32(-1), result)
	})
	
	t.Run("ShouldCapVeryLargeNegativeLength", func(t *testing.T) {
		result := LenToInt32(math.MinInt)
		
		assert.Equal(t, int32(math.MinInt32), result)
	})
	
	t.Run("ShouldWorkWithActualSliceLength", func(t *testing.T) {
		slice := make([]int, 100)
		result := LenToInt32(len(slice))
		
		assert.Equal(t, int32(100), result)
	})
	
	t.Run("ShouldWorkWithActualStringLength", func(t *testing.T) {
		str := "Hello, World!"
		result := LenToInt32(len(str))
		
		assert.Equal(t, int32(13), result)
	})
}

// Integration tests to verify functions work together
func TestIntegration(t *testing.T) {
	t.Run("ShouldWorkWithChainedConversions", func(t *testing.T) {
		// Test a realistic scenario: converting slice length to various types
		slice := make([]string, 50)
		length := len(slice)
		
		// Convert to uint64 (with error handling)
		uint64Result, err := IntToUInt64(length)
		assert.NoError(t, err)
		assert.Equal(t, uint64(50), uint64Result)
		
		// Convert to uint (with error handling)
		uintResult, err := IntToUInt(length)
		assert.NoError(t, err)
		assert.Equal(t, uint(50), uintResult)
		
		// Convert to int32 (with capping)
		int32Result := IntToInt32(length)
		assert.Equal(t, int32(50), int32Result)
		
		// Convert length directly to int32
		lenResult := LenToInt32(length)
		assert.Equal(t, int32(50), lenResult)
		assert.Equal(t, int32Result, lenResult)
	})
	
	t.Run("ShouldHandleLargeLengthsConsistently", func(t *testing.T) {
		largeLength := math.MaxInt32 + 1000
		
		// Both should cap at MaxInt32
		int32Result := IntToInt32(largeLength)
		lenResult := LenToInt32(largeLength)
		
		assert.Equal(t, int32(math.MaxInt32), int32Result)
		assert.Equal(t, int32(math.MaxInt32), lenResult)
		assert.Equal(t, int32Result, lenResult)
	})
	
	t.Run("ShouldHandleEdgeCasesConsistently", func(t *testing.T) {
		testCases := []int{
			0,
			1,
			-1,
			math.MaxInt32,
			math.MinInt32,
			math.MaxInt32 + 1,
			math.MinInt32 - 1,
		}
		
		for _, testCase := range testCases {
			int32Result := IntToInt32(testCase)
			lenResult := LenToInt32(testCase)
			
			// Both functions should produce the same result
			assert.Equal(t, int32Result, lenResult, "Results should be equal for input: %d", testCase)
		}
	})
}

// Benchmark tests
func BenchmarkIntToUInt64_Positive(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IntToUInt64(42)
	}
}

func BenchmarkIntToUInt64_Negative(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IntToUInt64(-42)
	}
}

func BenchmarkIntToUInt_Positive(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IntToUInt(42)
	}
}

func BenchmarkIntToUInt_Negative(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IntToUInt(-42)
	}
}

func BenchmarkInt64ToUint64_Positive(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Int64ToUint64(42)
	}
}

func BenchmarkInt64ToUint64_Negative(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Int64ToUint64(-42)
	}
}

func BenchmarkIntToInt32_Normal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IntToInt32(42)
	}
}

func BenchmarkIntToInt32_Overflow(b *testing.B) {
	value := math.MaxInt32 + 1000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IntToInt32(value)
	}
}

func BenchmarkLenToInt32_Normal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LenToInt32(42)
	}
}

func BenchmarkLenToInt32_Overflow(b *testing.B) {
	value := math.MaxInt32 + 1000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LenToInt32(value)
	}
}
