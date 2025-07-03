package signoz

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestSpanCountingProcessor(t *testing.T) {
	// Create a simple processor to wrap
	baseProcessor := sdktrace.NewSimpleSpanProcessor(nil)
	
	// Create span counting processor
	spanCounter := NewSpanCountingProcessor(baseProcessor, true)
	
	// Create a tracer provider with our counting processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanCounter),
	)
	
	// Set global tracer provider
	otel.SetTracerProvider(tp)
	
	// Create a tracer
	tracer := otel.Tracer("test-tracer")
	
	// Start a parent span
	ctx := context.Background()
	parentCtx, parentSpan := tracer.Start(ctx, "parent-span")
	
	// Start some child spans
	_, childSpan1 := tracer.Start(parentCtx, "child-span-1")
	_, childSpan2 := tracer.Start(parentCtx, "child-span-2")
	
	// Get the trace ID
	traceID := parentSpan.SpanContext().TraceID().String()
	
	// End child spans
	childSpan1.End()
	childSpan2.End()
	
	// Check span count before ending parent
	count := spanCounter.GetSpanCount(traceID)
	if count != 3 {
		t.Errorf("Expected 3 spans, got %d", count)
	}
	
	// End parent span
	parentSpan.End()
	
	// Give some time for processing
	time.Sleep(100 * time.Millisecond)
	
	// Check that the trace was cleaned up
	finalCount := spanCounter.GetSpanCount(traceID)
	if finalCount != 0 {
		t.Errorf("Expected 0 spans after cleanup, got %d", finalCount)
	}
	
	// Shutdown
	tp.Shutdown(context.Background())
}

func TestSpanCountingConfigDefault(t *testing.T) {
	cfg := &Config{}
	
	spanCountingConfig := cfg.GetSpanCountingConfig()
	if spanCountingConfig.Enabled != false {
		t.Errorf("Expected default Enabled to be false, got %v", spanCountingConfig.Enabled)
	}
	
	if spanCountingConfig.LogSpanCounts != false {
		t.Errorf("Expected default LogSpanCounts to be false, got %v", spanCountingConfig.LogSpanCounts)
	}
}

func TestSpanCountingConfigEnabled(t *testing.T) {
	cfg := &Config{
		SpanCountingConfig: &SpanCountingConfig{
			Enabled:       true,
			LogSpanCounts: true,
		},
	}
	
	spanCountingConfig := cfg.GetSpanCountingConfig()
	if spanCountingConfig.Enabled != true {
		t.Errorf("Expected Enabled to be true, got %v", spanCountingConfig.Enabled)
	}
	
	if spanCountingConfig.LogSpanCounts != true {
		t.Errorf("Expected LogSpanCounts to be true, got %v", spanCountingConfig.LogSpanCounts)
	}
}