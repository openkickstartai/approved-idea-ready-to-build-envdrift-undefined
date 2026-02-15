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

## Usage

```bash
# Scan current directory against .env.example
envdrift -dir . -env .env.example

# JSON output for CI pipelines
envdrift -dir ./src -env .env.example -format json
```

### Example Output

```
MISSING — referenced in code but not in env file (1):
  ✗ JWT_SECRET
      app.go:12
UNUSED — defined in env file but not in code (1):
  ? OLD_API_KEY
```

## Supported Languages

| Language | Patterns detected |
|---|---|
| Go | `os.Getenv("X")`, `os.LookupEnv("X")` |
| Python | `os.environ["X"]`, `os.environ.get("X")`, `os.getenv("X")` |
| JS/TS | `process.env.X`, `process.env["X"]` |

## CI Integration

```yaml
- name: EnvDrift check
  run: go run github.com/openkickstart/envdrift@latest -dir . -env .env.example
```

Exit code `1` when missing vars are found. Exit code `0` when clean.

## License

MIT
