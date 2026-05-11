# speedtestcli

Agentic-native internet speed test CLI. Tests against Cloudflare and Ookla/speedtest.net. Outputs structured JSON for programmatic consumption.

## Install

```bash
go install github.com/lroolle/speedtestcli@latest
```

Or build from source:

```bash
git clone https://github.com/lroolle/speedtestcli.git
cd speedtestcli
make build
```

## Usage

```bash
# Default: both backends in parallel, JSON output
speedtest

# Quick probe (~5s, latency + small download, no upload)
speedtest --quick

# Single backend
speedtest --backend=cloudflare
speedtest --backend=ookla

# Human-readable text output
speedtest --quick --format=text

# Streaming NDJSON (one event per line)
speedtest --format=ndjson

# Thorough test (sequential backends, per-phase timeouts)
speedtest --thorough
```

## Output

Default outputs JSON to stdout. Exit codes: 0=success, 1=network error, 2=invalid args, 3=timeout.

```bash
# Extract download speed
speedtest --quick --backend=cloudflare | jq '.download.bits_per_sec'

# Check if proxy is detected
speedtest --quick | jq '.proxy_detected'

# Get latency median
speedtest --quick | jq '.latency.stats.median_ms'
```

### Multi-backend (default)

```json
{
  "timestamp": "...",
  "duration_s": 12.5,
  "preset": "quick",
  "results": [
    {"backend": "cloudflare", "status": "ok", "granularity": "per-sample", ...},
    {"backend": "ookla", "status": "ok", "granularity": "aggregate", ...}
  ]
}
```

### Single backend

```json
{
  "id": "...",
  "status": "ok",
  "backend": "cloudflare",
  "granularity": "per-sample",
  "connection": {"client_ip": "...", "asn": 13335, "city": "Los Angeles", ...},
  "latency": {"samples": 5, "stats": {"median_ms": 12.3, ...}, "jitter_ms": 1.2},
  "download": {"bits_per_sec": 524288000, "samples": 8, ...},
  "upload": {"bits_per_sec": 0, "samples": 0, ...},
  "proxy_detected": {"https_proxy": "http://127.0.0.1:7890"},
  "errors": []
}
```

### Status field

- `"ok"` -- all phases completed
- `"partial"` -- some phases failed (e.g., download succeeded but upload timed out)
- `"failed"` -- could not complete any measurement

Partial results are always emitted. An agent can use whatever data was collected.

### Granularity field

- `"per-sample"` (Cloudflare) -- each TestStep produces individual samples with timing traces
- `"aggregate"` (Ookla) -- backend runs its own internal test plan, returns a single aggregate number

## Flags

```
--backend string     all (default), cloudflare, ookla
--format string      json (default), ndjson, text
--quick              Quick test (~5s, latency + small download)
--thorough           Thorough test (~4min, large payloads, sequential backends)
--timeout duration   Override test timeout
--base-url string    Override Cloudflare base URL
--no-upload          Skip upload tests
--no-download        Skip download tests
--verbose            Print structured log lines to stderr
--version            Print version
```

## Architecture

```
cmd/           CLI layer (cobra, flag parsing, output formatting)
pkg/speedtest/ Library (Backend interface, Runner, events, stats)
  cfbackend/   Cloudflare backend (own HTTP logic against speed.cloudflare.com)
  ooklabackend/ Ookla backend (wraps speedtest-go library)
internal/      Version injection, CLI utilities
```

## License

MIT
