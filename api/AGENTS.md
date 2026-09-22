# AGENTS.md — api/

Go door for PII module. Parent rules: root [`AGENTS.md`](../AGENTS.md).

## Role

Validate `POST /process`, rate-limit (default 1500), forward to compute,
map transport errors. **No** Redis, detectors, or mask logic here.

## Layout (target)

```
api/
  main.go
  go.mod
  internal/config/
  internal/handler/     # process, health
  internal/middleware/  # logging, ratelimit
  internal/forwarder/
  internal/ratelimit/
  internal/models/
```

## Commands

```bash
cd api && go test ./... && go vet ./...
gofmt -w .
```

## Boundaries

- Ownership track A: write only `api/**`.
- Frozen JSON: `payload`, `payload_id` → `result`.
- Never log payload/mask/original.
- No comments in code.
