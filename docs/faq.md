## Why not UUID?

- No ordering
- Larger storage
- Worse indexing

## Why not DB sequences?

- Central bottleneck
- No offline generation
- Poor global scaling

## Can versions be compared?

- Only within same version
- Cross-version comparison is undefined by design
