package snowflake

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Config{Version: Version0, NodeID: 100},
			wantErr: false,
		},
		{
			name:    "max node ID",
			config:  Config{Version: Version0, NodeID: 255},
			wantErr: false,
		},
		{
			name:    "invalid node ID",
			config:  Config{Version: Version0, NodeID: 256},
			wantErr: true,
		},
		{
			name:    "invalid version",
			config:  Config{Version: 99, NodeID: 100},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGenerator(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gen == nil {
				t.Error("NewGenerator() returned nil generator")
			}
		})
	}
}

func TestNextID_UniqueIDs(t *testing.T) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	const numIDs = 10000
	ids := make(map[uint64]bool, numIDs)

	for i := 0; i < numIDs; i++ {
		id, err := gen.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		if ids[id] {
			t.Fatalf("Duplicate ID generated: %d", id)
		}
		ids[id] = true
	}

	if len(ids) != numIDs {
		t.Errorf("Expected %d unique IDs, got %d", numIDs, len(ids))
	}
}

func TestNextID_Monotonic(t *testing.T) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	const numIDs = 1000
	var lastID uint64

	for i := 0; i < numIDs; i++ {
		id, err := gen.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		if i > 0 && id <= lastID {
			t.Fatalf("IDs not monotonically increasing: %d <= %d", id, lastID)
		}
		lastID = id
	}
}

func TestNextID_Concurrent(t *testing.T) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	const numGoroutines = 10
	const idsPerGoroutine = 1000
	const totalIDs = numGoroutines * idsPerGoroutine

	var wg sync.WaitGroup
	idsChan := make(chan uint64, totalIDs)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := gen.NextID()
				if err != nil {
					t.Errorf("Failed to generate ID: %v", err)
					return
				}
				idsChan <- id
			}
		}()
	}

	wg.Wait()
	close(idsChan)

	// Check for duplicates
	ids := make(map[uint64]bool, totalIDs)
	for id := range idsChan {
		if ids[id] {
			t.Fatalf("Duplicate ID generated in concurrent test: %d", id)
		}
		ids[id] = true
	}

	if len(ids) != totalIDs {
		t.Errorf("Expected %d unique IDs, got %d", totalIDs, len(ids))
	}
}

func TestDecode(t *testing.T) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 42})
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id, err := gen.NextID()
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	decoded, err := Decode(id)
	if err != nil {
		log.Fatal(err)
	}

	if decoded == nil {
		t.Fatal("Decode returned nil")
	}

	if decoded.Version != Version0 {
		t.Errorf("Expected version %d, got %d", Version0, decoded.Version)
	}

	if decoded.NodeID != 42 {
		t.Errorf("Expected node ID 42, got %d", decoded.NodeID)
	}

	// Verify timestamp is reasonable (within last second)
	now := time.Now()
	timeDiff := now.Sub(decoded.Time)
	if timeDiff < 0 || timeDiff > time.Second {
		t.Errorf("Decoded time seems incorrect: %v (diff: %v)", decoded.Time, timeDiff)
	}
}

func TestDecode_MultipleIDs(t *testing.T) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 100})
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	const numIDs = 100
	ids := make([]uint64, numIDs)

	for i := 0; i < numIDs; i++ {
		id, err := gen.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}
		ids[i] = id
	}

	for i, id := range ids {
		decoded, err := Decode(id)
		if err != nil {
			log.Fatal(err)
		}
		if decoded == nil {
			t.Fatalf("Failed to decode ID %d", id)
		}

		if decoded.NodeID != 100 {
			t.Errorf("ID %d: expected node ID 100, got %d", i, decoded.NodeID)
		}

		if decoded.Version != Version0 {
			t.Errorf("ID %d: expected version %d, got %d", i, Version0, decoded.Version)
		}
	}
}

func TestSequenceOverflow(t *testing.T) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Generate enough IDs to trigger sequence overflow
	// With 8 bits, we can generate 256 IDs per millisecond
	const numIDs = 3000
	ids := make([]uint64, numIDs)

	start := time.Now()
	for i := 0; i < numIDs; i++ {
		id, err := gen.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID at index %d: %v", i, err)
		}
		ids[i] = id
	}
	duration := time.Since(start)

	// Check all IDs are unique
	idSet := make(map[uint64]bool)
	for _, id := range ids {
		if idSet[id] {
			t.Fatalf("Duplicate ID found: %d", id)
		}
		idSet[id] = true
	}

	// Should take at least 1ms due to overflow
	if duration < time.Millisecond {
		t.Logf("Generated %d IDs in %v (should include wait time)", numIDs, duration)
	}
}

func TestMultipleGenerators_DifferentNodes(t *testing.T) {
	gen1, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		t.Fatalf("Failed to create generator 1: %v", err)
	}

	gen2, err := NewGenerator(Config{Version: Version0, NodeID: 2})
	if err != nil {
		t.Fatalf("Failed to create generator 2: %v", err)
	}

	const numIDs = 1000
	ids := make(map[uint64]bool, numIDs*2)

	// Generate from both generators
	for i := 0; i < numIDs; i++ {
		id1, err := gen1.NextID()
		if err != nil {
			t.Fatalf("Generator 1 failed: %v", err)
		}

		id2, err := gen2.NextID()
		if err != nil {
			t.Fatalf("Generator 2 failed: %v", err)
		}

		if ids[id1] {
			t.Fatalf("Duplicate ID from generator 1: %d", id1)
		}
		if ids[id2] {
			t.Fatalf("Duplicate ID from generator 2: %d", id2)
		}

		ids[id1] = true
		ids[id2] = true
	}

	if len(ids) != numIDs*2 {
		t.Errorf("Expected %d unique IDs, got %d", numIDs*2, len(ids))
	}
}

func TestVersionEncoding(t *testing.T) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id, err := gen.NextID()
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	// Extract version from top 3 bits
	version := Version((id >> 61) & 0x07)
	if version != Version0 {
		t.Errorf("Expected version %d encoded in ID, got %d", Version0, version)
	}
}

// Benchmark tests
func BenchmarkNextID(b *testing.B) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.NextID()
		if err != nil {
			b.Fatalf("Failed to generate ID: %v", err)
		}
	}
}

func BenchmarkNextID_Parallel(b *testing.B) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := gen.NextID()
			if err != nil {
				b.Fatalf("Failed to generate ID: %v", err)
			}
		}
	})
}

func BenchmarkDecode(b *testing.B) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	id, err := gen.NextID()
	if err != nil {
		b.Fatalf("Failed to generate ID: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(id)
	}
}

// Performance test to measure IDs per second
func TestPerformance_IDsPerSecond(t *testing.T) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	duration := 1 * time.Second
	start := time.Now()
	count := 0

	for time.Since(start) < duration {
		_, err := gen.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}
		count++
	}

	idsPerSecond := float64(count) / time.Since(start).Seconds()
	t.Logf("Generated %.0f IDs per second", idsPerSecond)

	// Should generate at least 100k IDs per second
	if idsPerSecond < 100000 {
		t.Logf("Warning: Low throughput (%.0f IDs/sec)", idsPerSecond)
	}
}

func TestPerformance_ConcurrentIDsPerSecond(t *testing.T) {
	gen, err := NewGenerator(Config{Version: Version0, NodeID: 1})
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	const numWorkers = 10
	duration := 1 * time.Second

	var wg sync.WaitGroup
	counts := make([]int, numWorkers)
	start := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			count := 0
			for time.Since(start) < duration {
				_, err := gen.NextID()
				if err != nil {
					t.Errorf("Worker %d failed: %v", workerID, err)
					return
				}
				count++
			}
			counts[workerID] = count
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	totalCount := 0
	for _, c := range counts {
		totalCount += c
	}

	idsPerSecond := float64(totalCount) / elapsed.Seconds()
	t.Logf("Generated %.0f IDs per second with %d concurrent workers", idsPerSecond, numWorkers)
}
