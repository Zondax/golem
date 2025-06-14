package zobservability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionStatus_WhenValidStatuses_ShouldHaveCorrectValues(t *testing.T) {
	// Assert that transaction status constants have expected values
	assert.Equal(t, "ok", string(TransactionOK))
	assert.Equal(t, "error", string(TransactionError))
	assert.Equal(t, "cancelled", string(TransactionCancelled))
}

func TestWithTransactionTag_WhenAppliedToTransaction_ShouldNotPanic(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}
	option := WithTransactionTag("key", "value")

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyTransaction(transaction)
	})
}

func TestWithTransactionTag_WhenAppliedWithEmptyKey_ShouldNotPanic(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}
	option := WithTransactionTag("", "value")

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyTransaction(transaction)
	})
}

func TestWithTransactionTag_WhenAppliedWithEmptyValue_ShouldNotPanic(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}
	option := WithTransactionTag("key", "")

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyTransaction(transaction)
	})
}

func TestWithTransactionData_WhenAppliedToTransaction_ShouldNotPanic(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}
	data := map[string]interface{}{"test": "data"}
	option := WithTransactionData("key", data)

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyTransaction(transaction)
	})
}

func TestWithTransactionData_WhenAppliedWithNilData_ShouldNotPanic(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}
	option := WithTransactionData("key", nil)

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyTransaction(transaction)
	})
}

func TestWithTransactionData_WhenAppliedWithComplexData_ShouldNotPanic(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}
	complexData := map[string]interface{}{
		"string":  "value",
		"number":  42,
		"boolean": true,
		"array":   []string{"a", "b", "c"},
		"nested": map[string]interface{}{
			"inner": "value",
		},
	}
	option := WithTransactionData("complex", complexData)

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		option.ApplyTransaction(transaction)
	})
}

func TestTransactionOptionFunc_WhenImplementsTransactionOption_ShouldBeValidInterface(t *testing.T) {
	// Arrange
	var option TransactionOption = transactionOptionFunc(func(t Transaction) {
		t.SetTag("test", "value")
	})

	// Assert
	assert.NotNil(t, option)
	assert.Implements(t, (*TransactionOption)(nil), option)
}

func TestTransactionOptionFunc_WhenAppliedToTransaction_ShouldExecuteFunction(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}
	executed := false

	option := transactionOptionFunc(func(t Transaction) {
		executed = true
		t.SetTag("test", "value")
	})

	// Act
	option.ApplyTransaction(transaction)

	// Assert
	assert.True(t, executed)
}

func TestTransactionOptions_WhenMultipleOptionsApplied_ShouldNotPanic(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}
	data := map[string]interface{}{"key": "value"}

	options := []TransactionOption{
		WithTransactionTag("tag1", "value1"),
		WithTransactionTag("tag2", "value2"),
		WithTransactionData("data1", data),
	}

	// Act & Assert - should not panic
	for _, option := range options {
		assert.NotPanics(t, func() {
			option.ApplyTransaction(transaction)
		})
	}
}

func TestWithTransactionTag_WhenUsedWithCommonTags_ShouldNotPanic(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}
	commonTags := map[string]string{
		TagOperation: "create-user",
		TagService:   "user-service",
		TagComponent: "user-repository",
		TagLayer:     LayerService,
		TagMethod:    "CreateUser",
	}

	var options []TransactionOption
	for key, value := range commonTags {
		options = append(options, WithTransactionTag(key, value))
	}

	// Act & Assert - should not panic
	for _, option := range options {
		assert.NotPanics(t, func() {
			option.ApplyTransaction(transaction)
		})
	}
}

func TestWithTransactionData_WhenUsedWithDifferentDataTypes_ShouldNotPanic(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}

	testCases := []struct {
		name string
		key  string
		data interface{}
	}{
		{"string_data", "string_key", "string_value"},
		{"int_data", "int_key", 42},
		{"bool_data", "bool_key", true},
		{"float_data", "float_key", 3.14},
		{"slice_data", "slice_key", []string{"a", "b", "c"}},
		{"map_data", "map_key", map[string]string{"nested": "value"}},
		{"nil_data", "nil_key", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			option := WithTransactionData(tc.key, tc.data)

			// Act & Assert - should not panic
			assert.NotPanics(t, func() {
				option.ApplyTransaction(transaction)
			})
		})
	}
}

func TestTransactionOption_WhenImplementedAsInterface_ShouldWork(t *testing.T) {
	// Arrange
	transaction := &noopTransaction{}

	// Test that all option types implement the interface correctly
	options := []TransactionOption{
		WithTransactionTag("key", "value"),
		WithTransactionData("data", "value"),
	}

	// Act & Assert
	for i, option := range options {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			assert.Implements(t, (*TransactionOption)(nil), option)
			assert.NotPanics(t, func() {
				option.ApplyTransaction(transaction)
			})
		})
	}
}

func TestTransactionOptions_WhenUsedInRealScenario_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()

	// Act - simulate real usage scenario
	transaction := observer.StartTransaction(ctx, "test-transaction",
		WithTransactionTag(TagLayer, LayerService),
		WithTransactionTag(TagComponent, "user-service"),
		WithTransactionData("user_id", "123"),
	)

	// Assert
	assert.NotNil(t, transaction)
	assert.Implements(t, (*Transaction)(nil), transaction)

	// Should not panic when setting additional properties
	assert.NotPanics(t, func() {
		transaction.SetName("updated-name")
		transaction.SetTag("additional", "tag")
		transaction.SetData("additional", "data")
	})

	// Should not panic when starting child span
	var childSpan Span
	assert.NotPanics(t, func() {
		childSpan = transaction.StartChild("child-operation")
	})
	assert.NotNil(t, childSpan)
	assert.Implements(t, (*Span)(nil), childSpan)

	// Should not panic when finishing
	assert.NotPanics(t, func() {
		childSpan.Finish()
		transaction.Finish(TransactionOK)
	})
}

func TestTransactionOptions_WhenChainedOperations_ShouldWorkCorrectly(t *testing.T) {
	// Arrange
	observer := NewNoopObserver()
	ctx := context.Background()

	// Act - complex transaction workflow
	transaction := observer.StartTransaction(ctx, "complex-transaction",
		WithTransactionTag("service", "user-service"),
		WithTransactionData("request_id", "123"),
	)

	// Add more properties after creation
	transaction.SetName("updated-transaction-name")
	transaction.SetTag("environment", "test")
	transaction.SetData("user_count", 42)

	// Create child spans
	span1 := transaction.StartChild("database-query",
		WithSpanTag("table", "users"),
		WithSpanData("query", "SELECT * FROM users"),
	)

	span2 := transaction.StartChild("cache-lookup",
		WithSpanTag("cache_type", "redis"),
	)

	// Assert
	assert.NotNil(t, transaction)
	assert.NotNil(t, span1)
	assert.NotNil(t, span2)

	// Should not panic when finishing in order
	assert.NotPanics(t, func() {
		span1.Finish()
		span2.Finish()
		transaction.Finish(TransactionOK)
	})
}
