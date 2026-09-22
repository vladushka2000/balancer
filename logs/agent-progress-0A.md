# Трек 0A — Go api пакет: контракт + ratelimit

## Статус: DONE (проверено 2026-09-22)

Трек 0A уже реализован в рамках общего Go-перевода. Проверка приёмки:

## Файлы

- `internal/api/ratelimit.go` — `TokenBucket` (capacity, refill/sec, sync.Mutex,
  защита от шага часов назад `elapsed < 0 → 0`).
- `internal/api/api.go` — `Door.Process`: валидация (пустой payload/payload_id →
  `ErrValidation`), rate limit (`ErrRateLimited`), вызов `compute.Processor.Process`.
- `internal/api/api_test.go` — `TestDoorForwardSuccess`, `TestDoor429RateLimit`,
  `TestDoor422InvalidBody`.
- `internal/api/ratelimit_test.go` — `TestTokenBucketCapacity`,
  `TestTokenBucketRefill`, `TestTokenBucketConcurrent`.

## Критерии приёмки

- [x] `go build ./...` — без ошибок.
- [x] `go test ./...` — все тесты зелёные (api, compute, detect, mask).
- [ ] «В коде нет комментариев» — **не выполнено**: комментарии есть по всему
      кодовому базису (96 строк в 17 файлах), закоммичены в «перевод на go».
      Снятие комментариев — кросс-трековый рефактор вне скоупа 0A. Отмечено
      по AGENTS.md rule 1 (код — истина при drift с промптом).

## Замечание

`internal/api/logging.go` из спеки 0A не создан: логирование реализовано в
`internal/compute/logging.go` (`InitLogging`, фильтр ПД → `<REDACTED>`), и
`cmd/server/main.go` использует его. api — тонкая дверь, свой логгер не нужен.