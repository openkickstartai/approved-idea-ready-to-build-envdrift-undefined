# EnvDrift

Fast, zero-dependency environment variable drift scanner. Catches mismatches between your code and `.env` files **before** runtime blows up with `undefined`.

Single static binary. Sub-millisecond on most repos.

## Install

```bash
go install github.com/openkickstart/envdrift@latest
```

Or build from source:

```bash
git clone https://github.com/openkickstart/envdrift.git
cd envdrift
go build -o envdrift .
```

## Quick Start

Install in one command:

```bash
go install github.com/openkickstart/envdrift@latest
```

**Scan a single directory** against the default `.env.example`:

```bash
envdrift -dir ./src
```

**Specify a custom `.env` file path**:

```bash
envdrift -dir . -env .env.production
```

**JSON output** for CI pipelines or programmatic consumption:

```bash
envdrift -dir ./src -env .env.example -format json
```

## Usage

```bash
# Scan current directory against .env.example
envdrift -dir . -env .env.example

# JSON output for CI pipelines
envdrift -dir ./src -env .env.example -format json
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-dir` | `.` | Root directory to scan for source files |
| `-env` | `.env.example` | Env definition file to compare against |
| `-format` | `text` | Output format: `text` or `json` |

### Example Output

```
MISSING — referenced in code but not in env file (1):
  ✗ JWT_SECRET
      app.go:12
UNUSED — defined in env file but not in code (1):
  ? OLD_API_KEY
```

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                    CLI (main.go)                     │
│         Parses flags: -dir, -env, -format            │
└──────────┬───────────────────────────────┬────────────┘
           │                               │
           ▼                               ▼
┌─────────────────────────┐   ┌─────────────────────────┐
│       Scanner           │   │      Env Parser         │
│     (scanner.go)        │   │     (scanner.go)        │
│                         │   │                         │
│  filepath.Walk +        │   │  ParseEnvFile() reads   │
│  regex patterns         │   │  KEY=VALUE definitions  │
│  → map[var][]Location   │   │  → map[var]string       │
└──────────┬──────────────┘   └──────────┬──────────────┘
           │                              │
           └──────────┬───────────────────┘
                      ▼
          ┌───────────────────────┐
          │   ComputeDrift()      │
          │                       │
          │  Code vars ∩ Env vars │
          │  → Missing + Unused   │
          └───────────┬───────────┘
                      ▼
          ┌───────────────────────┐
          │      Reporter         │
          │                       │
          │  report.Text() → text │
          │  report.JSON() → json │
          │                       │
          │  exit 1 if Missing >0 │
          │  exit 0 if clean      │
          └───────────────────────┘
```

## Supported Languages

| Language | Patterns detected |
|---|---|
| Go | `os.Getenv("X")`, `os.LookupEnv("X")` |
| Python | `os.environ["X"]`, `os.environ.get("X")`, `os.getenv("X")` |
| JS/TS | `process.env.X`, `process.env["X"]` |

## Comparison

| Feature | **EnvDrift** | dotenv-linter | envalid |
|---|---|---|---|
| Language | Go | Rust | JavaScript |
| Zero dependencies | ✅ stdlib only | ✅ | ❌ (npm) |
| Static analysis | ✅ regex, no runtime needed | ⚠️ lints `.env` files only | ❌ runtime validation |
| CI exit code | ✅ exit 1 on missing vars | ✅ | ❌ throws at boot |
| Multi-language scan | ✅ Go, Python, JS/TS | ❌ | ❌ JS only |
| Single static binary | ✅ | ✅ | ❌ |

## CI Integration

```yaml
- name: EnvDrift check
  run: go run github.com/openkickstart/envdrift@latest -dir . -env .env.example
```

Exit code `1` when missing vars are found. Exit code `0` when clean.

## License

MIT
