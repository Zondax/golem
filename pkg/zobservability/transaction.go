package zobservability

import "context"

// Transaction represents a logical operation that can contain multiple spans
type Transaction interface {
	Context() context.Context
	SetName(name string)
	SetTag(key, value string)
	SetData(key string, value interface{})
	StartChild(operation string, opts ...SpanOption) Span
	Finish(status TransactionStatus)
}

// TransactionStatus represents the final status of a transaction
type TransactionStatus string

const (
	// TransactionOK indicates a successful transaction
	TransactionOK TransactionStatus = "ok"
	// TransactionError indicates a failed transaction
	TransactionError TransactionStatus = "error"
	// TransactionCancelled indicates a cancelled transaction
	TransactionCancelled TransactionStatus = "cancelled"
)

// TransactionOption configures a transaction
type TransactionOption interface {
	ApplyTransaction(Transaction)
}

type transactionOptionFunc func(Transaction)

func (f transactionOptionFunc) ApplyTransaction(t Transaction) {
	f(t)
}

// WithTransactionTag adds a tag to the transaction
func WithTransactionTag(key, value string) TransactionOption {
	return transactionOptionFunc(func(t Transaction) {
		t.SetTag(key, value)
	})
}

// WithTransactionData adds data to the transaction
func WithTransactionData(key string, value interface{}) TransactionOption {
	return transactionOptionFunc(func(t Transaction) {
		t.SetData(key, value)
	})
}
