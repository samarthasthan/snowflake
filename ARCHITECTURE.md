# Architecture & Guarantees

## Core Principles

1. IDs are generated locally
2. No shared state across nodes
3. Time is authoritative but guarded
4. Layouts are versioned forever
5. Old IDs are never invalidated

## Guarantees

| Property | Guaranteed |
|-------|-----------|
| Global uniqueness | Yes |
| Monotonic per node | Yes |
| Sortable by time | Within same version |
| Offline generation | Yes |
| Multi-region safe | Via version upgrades |

## Non-Goals

- Cross-version strict ordering
- Central coordination
- Database-generated IDs

These are intentional tradeoffs.
