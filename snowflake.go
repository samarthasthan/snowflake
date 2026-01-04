package snowflake

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Version defines the layout version for ID generation
type Version uint8

const (
	// Version0 layout: [3 bits version][40 bits time][10 bits node][11 bits sequence]
	// Time unit: milliseconds, Epoch: 2026-01-01T00:00:00Z
	Version0 Version = 0
)

var (
	ErrInvalidNodeID     = errors.New("node ID out of range for version")
	ErrInvalidVersion    = errors.New("unsupported version")
	ErrClockRollback     = errors.New("clock moved backwards")
	ErrSequenceExhausted = errors.New("sequence exhausted for current millisecond")
)

// VersionLayout defines the bit layout and constraints for a version
type VersionLayout struct {
	Version      Version
	VersionBits  uint8
	TimeBits     uint8
	NodeBits     uint8
	SequenceBits uint8
	TimeUnit     time.Duration
	Epoch        time.Time
	MaxNodeID    uint64
	MaxSequence  uint64
	MaxTimestamp uint64
}

// Version layouts registry
var versionLayouts = map[Version]*VersionLayout{
	Version0: {
		Version:      Version0,
		VersionBits:  3,
		TimeBits:     45,
		NodeBits:     8,
		SequenceBits: 8,
		TimeUnit:     time.Millisecond,
		Epoch:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		MaxNodeID:    (1 << 8) - 1,  // 255
		MaxSequence:  (1 << 8) - 1,  // 255
		MaxTimestamp: (1 << 45) - 1, // ~1,118 years
	},
}

// Config holds generator configuration
type Config struct {
	Version Version
	NodeID  uint64
}

// Generator is a thread-safe Snowflake ID generator
type Generator struct {
	mu            sync.Mutex
	layout        *VersionLayout
	nodeID        uint64
	lastTimestamp uint64
	sequence      uint64

	// Bit shift positions for encoding
	versionShift uint8
	timeShift    uint8
	nodeShift    uint8
}

// DecodedID contains the components of a decoded Snowflake ID
type DecodedID struct {
	Version   Version
	Timestamp uint64
	NodeID    uint64
	Sequence  uint64
	Time      time.Time
}

// NewGenerator creates a new Snowflake ID generator
func NewGenerator(cfg Config) (*Generator, error) {
	layout, ok := versionLayouts[cfg.Version]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrInvalidVersion, cfg.Version)
	}

	if cfg.NodeID > layout.MaxNodeID {
		return nil, fmt.Errorf("%w: %d (max: %d)", ErrInvalidNodeID, cfg.NodeID, layout.MaxNodeID)
	}

	// Calculate bit shifts for encoding
	// Layout from MSB to LSB: [version][time][node][sequence]
	sequenceBits := layout.SequenceBits
	nodeBits := layout.NodeBits
	timeBits := layout.TimeBits
	// versionBits := uint8(3) // Always 3 bits for version

	g := &Generator{
		layout:        layout,
		nodeID:        cfg.NodeID,
		lastTimestamp: 0,
		sequence:      0,
		versionShift:  sequenceBits + nodeBits + timeBits,
		timeShift:     sequenceBits + nodeBits,
		nodeShift:     sequenceBits,
	}

	return g, nil
}

// NextID generates the next unique ID
func (g *Generator) NextID() (uint64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := g.currentTimestamp()

	if timestamp > g.layout.MaxTimestamp {
		return 0, errors.New("timestamp overflow for version")
	}

	// Handle clock rollback
	if timestamp < g.lastTimestamp {
		for timestamp < g.lastTimestamp {
			time.Sleep(100 * time.Microsecond)
			timestamp = g.currentTimestamp()
		}
	}

	// Same millisecond - increment sequence
	if timestamp == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & g.layout.MaxSequence

		// Sequence overflow - wait for next millisecond
		if g.sequence == 0 {
			timestamp = g.waitNextTimestamp(timestamp)
		}
	} else {
		// New millisecond - reset sequence
		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	// Encode ID: [version][timestamp][nodeID][sequence]
	id := (uint64(g.layout.Version) << g.versionShift) |
		(timestamp << g.timeShift) |
		(g.nodeID << g.nodeShift) |
		g.sequence

	return id, nil
}

func Decode(id uint64) (*DecodedID, error) {
	version, layout := extractVersion(id)
	if layout == nil {
		return nil, ErrInvalidVersion
	}

	timeShift := layout.SequenceBits + layout.NodeBits
	nodeShift := layout.SequenceBits

	timestamp := (id >> timeShift) & layout.MaxTimestamp
	nodeID := (id >> nodeShift) & layout.MaxNodeID
	sequence := id & layout.MaxSequence

	actualTime := layout.Epoch.Add(time.Duration(timestamp) * layout.TimeUnit)

	return &DecodedID{
		Version:   version,
		Timestamp: timestamp,
		NodeID:    nodeID,
		Sequence:  sequence,
		Time:      actualTime,
	}, nil
}

// currentTimestamp returns the current timestamp relative to epoch
func (g *Generator) currentTimestamp() uint64 {
	elapsed := time.Since(g.layout.Epoch)
	return uint64(elapsed / g.layout.TimeUnit)
}

// waitNextTimestamp waits until the next millisecond
func (g *Generator) waitNextTimestamp(lastTimestamp uint64) uint64 {
	timestamp := g.currentTimestamp()
	for timestamp <= lastTimestamp {
		time.Sleep(100 * time.Microsecond)
		timestamp = g.currentTimestamp()
	}
	return timestamp
}

// String returns a formatted representation of the decoded ID
func (d *DecodedID) String() string {
	return fmt.Sprintf("Version: %d, Time: %s, NodeID: %d, Sequence: %d",
		d.Version, d.Time.Format(time.RFC3339Nano), d.NodeID, d.Sequence)
}

func extractVersion(id uint64) (Version, *VersionLayout) {
	v := Version((id >> 61) & 0x07)
	layout := versionLayouts[v]
	return v, layout
}
