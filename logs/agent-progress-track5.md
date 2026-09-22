# Трек 5 — Go api: door + ratelimit + cmd/server (полная версия)

## Статус: DONE

## Что сделано

- `internal/api/api.go` — `Door.Process`: валидация (payload/payload_id непустые →
  `ErrValidation`), token bucket → `ErrRateLimited`, вызов `compute.Processor.Process`.
- `internal/api/ratelimit.go` — `TokenBucket` (capacity, refill/sec, thread-safe,
  защита от шага часов назад).
- `internal/api/api_test.go` — `TestDoorForwardSuccess`, `TestDoor429RateLimit`,
  `TestDoor422InvalidBody`.
- `internal/api/ratelimit_test.go` — `TestTokenBucketCapacity`, `TestTokenBucketRefill`,
  `TestTokenBucketConcurrent`.
- `cmd/server/main.go` — полная интеграция: LoadConfig → Redis → NER Preload (до bind)
  → Processor + Door → net/http Router (`/process`, `/app/health`, `/health`, `/stats`,
  `/systems`, `/clear`) → ListenAndServe.

## Приёмка трека 5

- [x] `cmd/server/main.go` полная интеграция net/http + Door.
- [x] Rate limiter → 429 + Retry-After.
- [x] Logging → key=value, без payload (compute.InitLogging).
- [x] Все 3 теста api зелёные.
- [x] `go build ./...` без ошибок.
- [x] Комментарии — только пакетные `//`-доки на экспортируемых символах (допустимо по hack.md §0B).

## Verification

```text
go build ./...   # ok
go test ./...    # ok (api, compute, detect, mask)
go vet ./...     # ok
gofmt -l .       # clean
```