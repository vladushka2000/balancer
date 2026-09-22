# Трек 4 — Go compute: pipeline + processor + admin

## Статус: DONE (проверено 2026-09-22)

Трек 4 уже реализован в рамках общего Go-перевода. Проверка приёмки:

## Файлы

- `internal/compute/pipeline.go` — `Pipeline.Process`: structural.DetectAll →
  ner.Detect → context.Filter → MergeSpans → masker.Apply → (masked, types).
- `internal/compute/processor.go` — `Processor.Process`: idempotency lookup
  (без семафора) → на miss семафор → pipeline → store.Put синхронно до 200 →
  stats.Record. Демаска/ретрай не берут семафор.
- `internal/compute/stats.go` — `Stats`: mean/p50/p95/p99, mask_ok, demask_ok,
  count_429, detections_by_type, requests_total.
- `internal/compute/logging.go` — slog key=value, фильтр payload/mask/original →
  `<REDACTED>`.
- `internal/compute/repo.go` — `Repo`: SaveSystem/GetSystem/ListSystems/
  BumpConfigEpoch/GetControlEpoch, TTL-кэш 2с, инвалидация при save.
- `cmd/server/main.go` — полная версия: LoadConfig → Redis → NER.Preload (до
  bind) → Processor + Door → net/http Router (/process, /app/health, /health,
  /stats, /systems, /clear) → ListenAndServe.
- Тесты: `pipeline_test.go` (8), `processor_test.go` (3) — всего 11.

## Критерии приёмки

- [x] `pipeline.go` оркестрирует detect→context→merge→mask.
- [x] `processor.go`: lookup без семафора; put синхронный до 200.
- [x] `stats.go` считает mean/p50/p95/p99, mask_ok, demask_ok, count_429.
- [x] `logging.go` фильтрует ПД из логов.
- [x] `cmd/server/main.go` инициализирует все зависимости.
- [x] Все 11 тестов зелёные (`go test ./internal/compute/ -run 'TestPipeline|TestProcessor'`).
- [x] `go build ./...`, `go vet ./...`, `gofmt -l .` — чисто.
- [x] Комментарии — только пакетные `//`-доки (допустимы по §0B); закомментированного
      кода нет.

## Замечание

`cmd/server/main.go` уже содержит полную интеграцию Door + TokenBucket + handlers
(трек 5), т.к. писался в рамках общего Go-перевода. `/process` использует
default-policy; admin `/systems`, `/stats`, `/clear` доступны.