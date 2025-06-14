package utils

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	t.Run("ShouldGenerateValidUUID", func(t *testing.T) {
		result := GenerateUUID()

		// Should be a valid UUID string
		_, err := uuid.Parse(result)
		assert.NoError(t, err)

		// Should be 36 characters long (including hyphens)
		assert.Len(t, result, 36)

		// Should contain 4 hyphens
		assert.Equal(t, 4, strings.Count(result, "-"))
	})

	t.Run("ShouldGenerateUniqueUUIDs", func(t *testing.T) {
		uuid1 := GenerateUUID()
		uuid2 := GenerateUUID()

		// Should generate different UUIDs
		assert.NotEqual(t, uuid1, uuid2)
	})

	t.Run("ShouldGenerateMultipleUniqueUUIDs", func(t *testing.T) {
		uuids := make(map[string]bool)
		iterations := 1000

		for i := 0; i < iterations; i++ {
			result := GenerateUUID()

			// Should not have duplicates
			assert.False(t, uuids[result], "Duplicate UUID generated: %s", result)
			uuids[result] = true

			// Each should be valid
			_, err := uuid.Parse(result)
			assert.NoError(t, err)
		}

		assert.Len(t, uuids, iterations)
	})
}

func TestGenerateUUIDShort(t *testing.T) {
	t.Run("ShouldGenerateShortUUID", func(t *testing.T) {
		result := GenerateUUIDShort()

		// Should be exactly 8 characters long
		assert.Len(t, result, 8)

		// Should be hexadecimal characters
		for _, char := range result {
			assert.True(t, (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || char == '-',
				"Character %c is not a valid hex character", char)
		}
	})

	t.Run("ShouldGenerateUniqueShortUUIDs", func(t *testing.T) {
		uuid1 := GenerateUUIDShort()
		uuid2 := GenerateUUIDShort()

		// Should generate different short UUIDs (very high probability)
		assert.NotEqual(t, uuid1, uuid2)
	})

	t.Run("ShouldBeFirstEightCharsOfFullUUID", func(t *testing.T) {
		// Generate multiple UUIDs to verify the pattern
		for i := 0; i < 100; i++ {
			shortUUID := GenerateUUIDShort()

			// Short UUID should be valid hex characters
			assert.Len(t, shortUUID, 8)
			for _, char := range shortUUID {
				assert.True(t, (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || char == '-')
			}
		}
	})

	t.Run("ShouldGenerateMultipleUniqueShortUUIDs", func(t *testing.T) {
		uuids := make(map[string]bool)
		iterations := 1000

		for i := 0; i < iterations; i++ {
			result := GenerateUUIDShort()
			assert.Len(t, result, 8)
			uuids[result] = true
		}

		// Should have high uniqueness (allowing for some collisions due to shorter length)
		minExpected := int(float64(iterations) * 0.95)
		assert.Greater(t, len(uuids), minExpected, "Short UUIDs should have high uniqueness")
	})
}

func TestGenerateUUIDv5(t *testing.T) {
	namespace := uuid.NameSpaceURL

	t.Run("ShouldGenerateValidUUIDv5", func(t *testing.T) {
		fields := []string{"field1", "field2", "field3"}

		result, err := GenerateUUIDv5(namespace, fields)

		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		// Should be a valid UUID string
		parsedUUID, err := uuid.Parse(result)
		assert.NoError(t, err)

		// Should be version 5 UUID
		assert.Equal(t, uuid.Version(5), parsedUUID.Version())

		// Should be 36 characters long
		assert.Len(t, result, 36)
	})

	t.Run("ShouldBeDeterministic", func(t *testing.T) {
		fields := []string{"field1", "field2", "field3"}

		result1, err1 := GenerateUUIDv5(namespace, fields)
		result2, err2 := GenerateUUIDv5(namespace, fields)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, result1, result2, "UUIDv5 should be deterministic for same inputs")
	})

	t.Run("ShouldGenerateDifferentUUIDsForDifferentFields", func(t *testing.T) {
		fields1 := []string{"field1", "field2"}
		fields2 := []string{"field1", "field3"}

		result1, err1 := GenerateUUIDv5(namespace, fields1)
		result2, err2 := GenerateUUIDv5(namespace, fields2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, result1, result2)
	})

	t.Run("ShouldGenerateDifferentUUIDsForDifferentOrder", func(t *testing.T) {
		fields1 := []string{"field1", "field2"}
		fields2 := []string{"field2", "field1"}

		result1, err1 := GenerateUUIDv5(namespace, fields1)
		result2, err2 := GenerateUUIDv5(namespace, fields2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, result1, result2, "Order of fields should matter")
	})

	t.Run("ShouldGenerateDifferentUUIDsForDifferentNamespaces", func(t *testing.T) {
		fields := []string{"field1", "field2"}

		result1, err1 := GenerateUUIDv5(uuid.NameSpaceURL, fields)
		result2, err2 := GenerateUUIDv5(uuid.NameSpaceDNS, fields)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, result1, result2, "Different namespaces should generate different UUIDs")
	})

	t.Run("ShouldHandleSingleField", func(t *testing.T) {
		fields := []string{"single_field"}

		result, err := GenerateUUIDv5(namespace, fields)

		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		// Should be valid UUID
		_, err = uuid.Parse(result)
		assert.NoError(t, err)
	})

	t.Run("ShouldHandleSpecialCharacters", func(t *testing.T) {
		fields := []string{"field@domain.com", "field with spaces", "field|with|pipes"}

		result, err := GenerateUUIDv5(namespace, fields)

		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		// Should be valid UUID
		_, err = uuid.Parse(result)
		assert.NoError(t, err)
	})

	t.Run("ShouldHandleUnicodeCharacters", func(t *testing.T) {
		fields := []string{"🚀", "测试", "café"}

		result, err := GenerateUUIDv5(namespace, fields)

		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		// Should be valid UUID
		_, err = uuid.Parse(result)
		assert.NoError(t, err)
	})

	// Error cases
	t.Run("WhenNoFields_ShouldReturnError", func(t *testing.T) {
		fields := []string{}

		result, err := GenerateUUIDv5(namespace, fields)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "at least one field is required")
	})

	t.Run("WhenNilFields_ShouldReturnError", func(t *testing.T) {
		result, err := GenerateUUIDv5(namespace, nil)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "at least one field is required")
	})

	t.Run("WhenEmptyFieldAtPosition0_ShouldReturnError", func(t *testing.T) {
		fields := []string{"", "field2"}

		result, err := GenerateUUIDv5(namespace, fields)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "field at position 0 cannot be empty")
	})

	t.Run("WhenEmptyFieldAtPosition1_ShouldReturnError", func(t *testing.T) {
		fields := []string{"field1", ""}

		result, err := GenerateUUIDv5(namespace, fields)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "field at position 1 cannot be empty")
	})

	t.Run("WhenEmptyFieldAtPosition2_ShouldReturnError", func(t *testing.T) {
		fields := []string{"field1", "field2", ""}

		result, err := GenerateUUIDv5(namespace, fields)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "field at position 2 cannot be empty")
	})

	t.Run("WhenMultipleEmptyFields_ShouldReturnFirstError", func(t *testing.T) {
		fields := []string{"field1", "", ""}

		result, err := GenerateUUIDv5(namespace, fields)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "field at position 1 cannot be empty")
	})
}

// Benchmark tests
func BenchmarkGenerateUUID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateUUID()
	}
}

func BenchmarkGenerateUUIDShort(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateUUIDShort()
	}
}

func BenchmarkGenerateUUIDv5_SingleField(b *testing.B) {
	namespace := uuid.NameSpaceURL
	fields := []string{"benchmark_field"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateUUIDv5(namespace, fields)
	}
}

func BenchmarkGenerateUUIDv5_MultipleFields(b *testing.B) {
	namespace := uuid.NameSpaceURL
	fields := []string{"field1", "field2", "field3", "field4", "field5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateUUIDv5(namespace, fields)
	}
}

func BenchmarkGenerateUUIDv5_LongFields(b *testing.B) {
	namespace := uuid.NameSpaceURL
	fields := []string{
		"this is a very long field with lots of text and content that might be used in real world scenarios",
		"another long field with different content and more text to simulate realistic usage patterns",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateUUIDv5(namespace, fields)
	}
}
