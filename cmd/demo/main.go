package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/samarthasthan/snowflake"
)

func main() {
	fmt.Println("=== Snowflake ID Generator Demo ===")

	// Example 1: Basic usage
	basicUsage()

	// Example 2: Concurrent generation
	concurrentGeneration()

	// Example 3: Decoding IDs
	decodingExample()

	// Example 4: Multiple nodes
	multipleNodesExample()

	// Example 5: Performance test
	performanceTest()

	// Example 6: Error handling
	errorHandlingExample()
}

func basicUsage() {
	fmt.Println("1. Basic Usage:")
	fmt.Println("   Creating generator with NodeID=1...")

	gen, err := snowflake.NewGenerator(snowflake.Config{
		Version: snowflake.Version0,
		NodeID:  1,
	})
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Generate a few IDs
	fmt.Println("   Generating 5 IDs:")
	for i := 0; i < 5; i++ {
		id, err := gen.NextID()
		if err != nil {
			log.Fatalf("Failed to generate ID: %v", err)
		}
		fmt.Printf("   - ID %d: %d\n", i+1, id)
	}
	fmt.Println()
}

func concurrentGeneration() {
	fmt.Println("2. Concurrent Generation:")
	fmt.Println("   Testing thread safety with 10 goroutines...")

	gen, err := snowflake.NewGenerator(snowflake.Config{
		Version: snowflake.Version0,
		NodeID:  2,
	})
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	const numGoroutines = 10
	const idsPerGoroutine = 100

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				_, err := gen.NextID()
				if err != nil {
					log.Printf("Worker %d failed: %v", workerID, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalIDs := numGoroutines * idsPerGoroutine
	fmt.Printf("   ✓ Generated %d IDs across %d goroutines in %v\n",
		totalIDs, numGoroutines, duration)
	fmt.Printf("   ✓ Throughput: %.0f IDs/sec\n\n",
		float64(totalIDs)/duration.Seconds())
}

func decodingExample() {
	fmt.Println("3. Decoding IDs:")

	gen, err := snowflake.NewGenerator(snowflake.Config{
		Version: snowflake.Version0,
		NodeID:  42,
	})
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	id, err := gen.NextID()
	if err != nil {
		log.Fatalf("Failed to generate ID: %v", err)
	}

	decoded, err := snowflake.Decode(id)
	if err != nil {
		log.Fatal(err)
	}
	if decoded == nil {
		log.Fatal("Failed to decode ID")
	}

	fmt.Printf("   Original ID: %d\n", id)
	fmt.Printf("   Version:     %d\n", decoded.Version)
	fmt.Printf("   Timestamp:   %d\n", decoded.Timestamp)
	fmt.Printf("   Time:        %s\n", decoded.Time.Format(time.RFC3339Nano))
	fmt.Printf("   NodeID:      %d\n", decoded.NodeID)
	fmt.Printf("   Sequence:    %d\n", decoded.Sequence)
	fmt.Println()
}

func multipleNodesExample() {
	fmt.Println("4. Multiple Nodes:")
	fmt.Println("   Creating generators for different nodes...")

	generators := make([]*snowflake.Generator, 3)
	nodeIDs := []uint64{10, 20, 30}

	for i, nodeID := range nodeIDs {
		gen, err := snowflake.NewGenerator(snowflake.Config{
			Version: snowflake.Version0,
			NodeID:  nodeID,
		})
		if err != nil {
			log.Fatalf("Failed to create generator for node %d: %v", nodeID, err)
		}
		generators[i] = gen
	}

	fmt.Println("   Generating IDs from each node:")
	for i, gen := range generators {
		id, err := gen.NextID()
		if err != nil {
			log.Fatalf("Failed to generate ID from node %d: %v", nodeIDs[i], err)
		}

		decoded, err := snowflake.Decode(id)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   - Node %d: ID=%d, Decoded NodeID=%d\n",
			nodeIDs[i], id, decoded.NodeID)
	}
	fmt.Println()
}

func performanceTest() {
	fmt.Println("5. Performance Test:")

	gen, err := snowflake.NewGenerator(snowflake.Config{
		Version: snowflake.Version0,
		NodeID:  1,
	})
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Test 1: Sequential generation
	fmt.Println("   Testing sequential generation...")
	count := 0
	start := time.Now()
	duration := 1 * time.Second

	for time.Since(start) < duration {
		_, err := gen.NextID()
		if err != nil {
			log.Fatalf("Failed to generate ID: %v", err)
		}
		count++
	}

	idsPerSec := float64(count) / time.Since(start).Seconds()
	fmt.Printf("   ✓ Sequential: %.0f IDs/sec (%d IDs in %v)\n",
		idsPerSec, count, time.Since(start))

	// Test 2: Check for duplicates in high-volume generation
	fmt.Println("   Testing duplicate detection (generating 100k IDs)...")
	ids := make(map[uint64]bool, 100000)
	duplicates := 0

	start = time.Now()
	for i := 0; i < 100000; i++ {
		id, err := gen.NextID()
		if err != nil {
			log.Fatalf("Failed to generate ID: %v", err)
		}

		if ids[id] {
			duplicates++
		}
		ids[id] = true
	}

	fmt.Printf("   ✓ Generated 100,000 unique IDs in %v\n", time.Since(start))
	fmt.Printf("   ✓ Duplicates found: %d\n", duplicates)

	if duplicates > 0 {
		fmt.Println("   ✗ WARNING: Duplicates detected!")
	}

	fmt.Println()
	fmt.Println("=== Demo Complete ===")
}

// Example of error handling
func errorHandlingExample() {
	fmt.Println("Error Handling Examples:")

	// Invalid node ID
	_, err := snowflake.NewGenerator(snowflake.Config{
		Version: snowflake.Version0,
		NodeID:  2000, // Too large
	})
	if err != nil {
		fmt.Printf("   Expected error for invalid node ID: %v\n", err)
	}

	// Invalid version
	_, err = snowflake.NewGenerator(snowflake.Config{
		Version: 99, // Unsupported
		NodeID:  1,
	})
	if err != nil {
		fmt.Printf("   Expected error for invalid version: %v\n", err)
	}
}
