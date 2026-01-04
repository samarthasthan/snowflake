# Snowflake ID Generator (Go)

A production-grade, versioned Snowflake ID generator designed for
long-lived distributed systems.

This library prioritizes:
- Global uniqueness
- Monotonic ordering (per node)
- Zero coordination
- Versioned evolution
- Decades-long lifespan

## Current Layout (v0)

[3 bits version][40 bits time(ms)][10 bits node][11 bits sequence]

- Max nodes: 1024
- Max IDs per node: ~2M/sec
- Time range: ~34.8 years
- Epoch: 2026-01-01T00:00:00Z

## Design Philosophy

- IDs are generated in the application, never the database
- No central coordination or network calls
- Clock rollback safe
- Future layouts supported via versioning

## Status

✅ Production ready  
✅ Race-safe  
✅ Benchmarked  
✅ Version-safe  

See ARCHITECTURE.md and VERSIONING.md for details.
