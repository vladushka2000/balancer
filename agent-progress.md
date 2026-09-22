# Прогресс

Индекс. Параллельные треки пишут в `logs/agent-progress-<track>.md`.
`feature_list.json` обновляет только glue/человек.

## Текущий фокус

Полный перевод на Go: api ↔ compute как пакеты одного модуля (вызовы функций,
не HTTP). Rust compute удалён.

## Сделано (недавнее)

### Синхронизация plan.md/hack.md с origin/reqs (2026-09-22)

- Входящие изменения `origin/reqs` для `plan.md` перенесены и согласованы с
  архитектурой «один Go-сервис, два пакета» (без gateway).
- `plan.md`: добавлены «Стек (зафиксировано)» (Go-only), инвариант
  `Go-only runtime`, шаг M0, §5.5 «Миграция с Rust», критерий Go-only.
- `hack.md` стал **основным планом**: заголовок «Основной план», стек Go-only,
  §0B «Миграция с Rust». Подробно описывает все фичи из `plan.md`.

### Go-перевод (2026-09-22)

- [`hack.md`](hack.md) переписан: Go-пакеты `internal/api` + `internal/compute`, HTTP-слой `cmd/server`.
- [`README.md`](README.md), [`AGENTS.md`](AGENTS.md), [`plan.md`](plan.md), [`feature_list.json`](feature_list.json), [`docs/agents/OWNERSHIP.md`](docs/agents/OWNERSHIP.md) обновлены под Go-only.
- Rust `compute/` удалён; `api/` (старый) удалён.
- Реализовано:
  - `internal/models` — замороженный контракт (DTO).
  - `internal/api` — Door (validate → token bucket → compute), TokenBucket.
  - `internal/compute` — detect (structural/ner/context/merge), mask (rules/masker), store (LRU+TTL+Redis write-through, AES-GCM), repo, semaphore, stats, pipeline, processor.
  - `cmd/server` — HTTP: `/process`, `/app/health`, `/systems`, `/stats`, `/clear`.

### Verification

```text
go build ./...   # ok
go test ./...    # ok (api, compute, detect, mask)
go vet ./...     # ok
gofmt -l .       # clean
```

Smoke-test (Redis на :6382): mask → demask roundtrip 100%, `mask_ok == demask_ok`, 422 на пустое тело, `/systems` + `/clear` работают.

## Дальше

1. `demo/selfcheck.py` + `demo/load.py` (разгон 330→1000).
2. `docker-compose.yml` + `Dockerfile`.
3. `scripts/pack.sh` + README ≤5 предложений.

## Блокеры

Нет.