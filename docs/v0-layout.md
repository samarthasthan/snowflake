# Snowflake Layout v0

Bit layout (MSB → LSB):

[ version | time | node | sequence ]

| Field    | Bits | Description             |
| -------- | ---- | ----------------------- |
| Version  | 3    | Layout version          |
| Time     | 45   | ms since epoch          |
| Node     | 8    | Generator (Node) ID     |
| Sequence | 8    | Per-ms counter per node |

Epoch:
2026-01-01T00:00:00Z

Max throughput:

- 256 IDs/ms/node
- ~65M IDs/sec globally

Time coverage:

- ~1,118 years
