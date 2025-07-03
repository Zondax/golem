package signoz

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/zondax/golem/pkg/logger"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SpanCountingProcessor wraps another SpanProcessor to count spans per trace
type SpanCountingProcessor struct {
	processor  trace.SpanProcessor
	spanCounts map[string]*int64 // traceID -> count
	mutex      sync.RWMutex
	logEnabled bool
}

// NewSpanCountingProcessor creates a new span counting processor
func NewSpanCountingProcessor(processor trace.SpanProcessor, logEnabled bool) *SpanCountingProcessor {
	return &SpanCountingProcessor{
		processor:  processor,
		spanCounts: make(map[string]*int64),
		logEnabled: logEnabled,
	}
}

// OnStart is called when a span starts
func (s *SpanCountingProcessor) OnStart(parent context.Context, span trace.ReadWriteSpan) {
	traceID := span.SpanContext().TraceID().String()

	s.mutex.Lock()
	counter, exists := s.spanCounts[traceID]
	if !exists {
		counter = new(int64)
		s.spanCounts[traceID] = counter
	}
	s.mutex.Unlock()

	count := atomic.AddInt64(counter, 1)

	if s.logEnabled {
		logger.GetLoggerFromContext(parent).Infof("Span started - TraceID: %s, SpanID: %s, Count: %d", 
			traceID, span.SpanContext().SpanID().String(), count)
	}

	// Call the wrapped processor
	s.processor.OnStart(parent, span)
}

// OnEnd is called when a span ends
func (s *SpanCountingProcessor) OnEnd(span trace.ReadOnlySpan) {
	traceID := span.SpanContext().TraceID().String()

	s.mutex.RLock()
	counter, exists := s.spanCounts[traceID]
	s.mutex.RUnlock()

	if exists {
		count := atomic.LoadInt64(counter)

		// If this is the root span (no parent), log final count and cleanup
		if !span.Parent().IsValid() {
			logger.GetLoggerFromContext(context.Background()).Infof("Trace completed - TraceID: %s, SpanID: %s, Count: %d",
				traceID, span.SpanContext().SpanID().String(), count)

			// Cleanup the counter for this trace
			s.mutex.Lock()
			delete(s.spanCounts, traceID)
			s.mutex.Unlock()
		}
	}

	// Call the wrapped processor
	s.processor.OnEnd(span)
}

// Shutdown shuts down the processor
func (s *SpanCountingProcessor) Shutdown(ctx context.Context) error {
	return s.processor.Shutdown(ctx)
}

// ForceFlush forces a flush of the processor
func (s *SpanCountingProcessor) ForceFlush(ctx context.Context) error {
	return s.processor.ForceFlush(ctx)
}

// GetSpanCount returns the current span count for a trace
func (s *SpanCountingProcessor) GetSpanCount(traceID string) int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if counter, exists := s.spanCounts[traceID]; exists {
		return atomic.LoadInt64(counter)
	}
	return 0
}

// GetAllSpanCounts returns a copy of all current span counts
func (s *SpanCountingProcessor) GetAllSpanCounts() map[string]int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	counts := make(map[string]int64)
	for traceID, counter := range s.spanCounts {
		counts[traceID] = atomic.LoadInt64(counter)
	}
	return counts
}
