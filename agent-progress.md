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

### Трек 0B (compute-фундамент, 2026-09-22)

- `internal/compute` уже реализован (models/config/keys/store/ratelimit/pipeline/processor/stats/repo/logging + detect/mask).
- Добавлены недостающие тесты трека 0B:
  - `keys_test.go` — corrKey/systemKey/systemsSet/configEpochKey.
  - `ratelimit_test.go` — Semaphore capacity/wait-timeout/release.
  - `config_test.go` — DefaultConfig + LoadConfig из env.
- Контракт §7 живёт в `internal/models/models.go` (замороженный пакет), совпадает буквально.

### Трек 3 (mask + dicts, 2026-09-22)

- `internal/compute/mask/rules.go` — 16 per-type mask-функций + `MaskRules()` (passport, fio, phone, email, card, inn, snils, date, cvv, pin, postal_code, department_code, address, driver_license, default).
- `internal/compute/mask/masker.go` — `Masker.Apply` right-to-left, режимы `partial`/`redact`.
- Словари в `internal/compute/detect/dicts/` (embed): famous.txt (20), org_addresses.txt (13), markers.txt (16), months.txt (12).
- Тесты `rules_test.go` + `masker_test.go` — зелёные.

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