package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zondax/golem/pkg/zobservability"
	"github.com/zondax/golem/pkg/zobservability/factory"
)

func main() {
	// Example 1: Configure B3 propagation
	config := &zobservability.Config{
		Provider:    zobservability.ProviderSigNoz,
		Enabled:     true,
		Environment: "development",
		Release:     "1.0.0",
		Address:     "localhost:4317",
		SampleRate:  1.0,
		Propagation: zobservability.PropagationConfig{
			Formats: []string{zobservability.PropagationB3},
		},
	}

	// Create observer with B3 propagation
	observer, err := factory.NewObserver(config, "b3-example-service")
	if err != nil {
		log.Fatalf("Failed to create observer: %v", err)
	}
	defer observer.Close()

	// Start a transaction with B3 propagation
	ctx := context.Background()
	tx := observer.StartTransaction(ctx, "b3-example-transaction")
	defer tx.Finish(zobservability.TransactionOK)

	// Create spans that will propagate B3 headers
	ctx, span := observer.StartSpan(tx.Context(), "b3-example-span")
	defer span.Finish()

	// Add some work
	time.Sleep(100 * time.Millisecond)

	fmt.Println("B3 propagation example completed")

	// Example 2: Configure multiple propagation formats (B3 + W3C)
	fmt.Println("\n--- Multiple Propagation Formats Example ---")
	
	multiConfig := &zobservability.Config{
		Provider:    zobservability.ProviderSigNoz,
		Enabled:     true,
		Environment: "development",
		Release:     "1.0.0",
		Address:     "localhost:4317",
		SampleRate:  1.0,
		Propagation: zobservability.PropagationConfig{
			Formats: []string{
				zobservability.PropagationB3,
				zobservability.PropagationW3C,
			},
		},
	}

	multiObserver, err := factory.NewObserver(multiConfig, "multi-format-service")
	if err != nil {
		log.Fatalf("Failed to create multi-format observer: %v", err)
	}
	defer multiObserver.Close()

	// Start transaction with multiple propagation formats
	multiTx := multiObserver.StartTransaction(ctx, "multi-format-transaction")
	defer multiTx.Finish(zobservability.TransactionOK)

	ctx, multiSpan := multiObserver.StartSpan(multiTx.Context(), "multi-format-span")
	defer multiSpan.Finish()

	fmt.Println("Multiple propagation formats example completed")

	// Example 3: B3 Single Header format
	fmt.Println("\n--- B3 Single Header Example ---")
	
	singleConfig := &zobservability.Config{
		Provider:    zobservability.ProviderSigNoz,
		Enabled:     true,
		Environment: "development",
		Release:     "1.0.0",
		Address:     "localhost:4317",
		SampleRate:  1.0,
		Propagation: zobservability.PropagationConfig{
			Formats: []string{zobservability.PropagationB3Single},
		},
	}

	singleObserver, err := factory.NewObserver(singleConfig, "b3-single-service")
	if err != nil {
		log.Fatalf("Failed to create B3 single header observer: %v", err)
	}
	defer singleObserver.Close()

	// Start transaction with B3 single header
	singleTx := singleObserver.StartTransaction(ctx, "b3-single-transaction")
	defer singleTx.Finish(zobservability.TransactionOK)

	ctx, singleSpan := singleObserver.StartSpan(singleTx.Context(), "b3-single-span")
	defer singleSpan.Finish()

	fmt.Println("B3 single header example completed")
}