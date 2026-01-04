# Operational Guidelines

## Clock Synchronization

Required:

- chrony (recommended)
- NTP fallback acceptable

Generator behavior:

- Never generates IDs when clock moves backward
- Blocks until time catches up

## Node ID Assignment

Rules:

- Node ID must be unique per machine
- Node ID must be stable across restarts
- Node ID is NOT per pod

Recommended sources:

- StatefulSet ordinal
- VM metadata
- Config management

## Anti-Patterns

❌ Random node IDs  
❌ Node ID per container  
❌ Database-generated IDs  
❌ Clock skew ignored
