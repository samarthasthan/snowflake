# Snowflake Versioning Policy

This library uses explicit layout versions to allow
safe evolution over decades.

## Version 0 (Current)

Layout:
[3v][45t(ms)][8n][8s]

Epoch:
2026-01-01T00:00:00Z

Capacity:

- ~65M IDs/sec globally
- ~1,118 years

## Version Change Rules

A new version is introduced when:

- Time bits near exhaustion
- Generator/node count exceeded
- Higher per-node throughput required
- Region/datacenter encoding required

Old versions:

- Remain valid forever
- Are still decodable
- Are never reinterpreted
- Are never re-encoded with new semantics

## Version 1 (Planned)

Proposed layout:
[3v][41t(ms)][10n][10s]

Epoch:
TBD (expected ~2035–2060)

Capacity:

- ~1B IDs/sec globally
- ~69 years

Notes:

- Reduced time span vs Version 0
- Higher per-node and global throughput
- Suitable for sustained high-write workloads
- Backward-compatible via version decoding
