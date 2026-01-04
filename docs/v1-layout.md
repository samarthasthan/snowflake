# Snowflake Layout v1 (Planned)

Bit layout (MSB → LSB):

[ version | time | node | sequence ]

| Field    | Bits | Description             |
| -------- | ---- | ----------------------- |
| Version  | 3    | Layout version          |
| Time     | 41   | ms since epoch          |
| Node     | 10   | Generator (Node) ID     |
| Sequence | 10   | Per-ms counter per node |

Epoch:
TBD (expected ~2035–2060)

Capacity:

- 1,024 IDs/ms/node
- ~1B IDs/sec globally
- ~69 years of time coverage

Motivation:

- Higher sustained throughput
- More generator instances
- Simpler decoding than geo-split layouts
- Suitable for high-write core entities

Tradeoffs:

- Shorter time span vs v0
- Requires planned version rollover
- Still requires coordinated generators
