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

### Трек 4 (compute: pipeline + processor + admin, 2026-09-22)

- `internal/compute/pipeline.go` — `Pipeline.Process`: detect→context→merge→mask.
- `internal/compute/processor.go` — `Processor.Process`: lookup без семафора; put синхронно до 200.
- `internal/compute/stats.go` — mean/p50/p95/p99, mask_ok, demask_ok, count_429.
- `internal/compute/logging.go` — slog, фильтр payload/mask/original → `<REDACTED>`.
- `internal/compute/repo.go` — `Repo` (SaveSystem/GetSystem/ListSystems/epoch), TTL-кэш 2с.
- `cmd/server/main.go` — полная интеграция: Redis, NER.Preload до bind, Router (/process, /app/health, /stats, /systems, /clear).
- Тесты `pipeline_test.go` (8) + `processor_test.go` (3) — зелёные.
- Подробности: `logs/agent-progress-4.md`.

### Трек 3 (mask + dicts, 2026-09-22)

- `internal/compute/mask/rules.go` — 16 per-type mask-функций + `MaskRules()` (passport, fio, phone, email, card, inn, snils, date, cvv, pin, postal_code, department_code, address, driver_license, default).
- `internal/compute/mask/masker.go` — `Masker.Apply` right-to-left, режимы `partial`/`redact`.
- Словари в `internal/compute/detect/dicts/` (embed): famous.txt (20), org_addresses.txt (13), markers.txt (16), months.txt (12).
- Тесты `rules_test.go` + `masker_test.go` — зелёные.

### Трек 6 (интеграция: compose + selfcheck + OpenAPI + pack, 2026-09-22)

- `docker-compose.yml` — redis + server (Go), env `REDIS_URL`/`PII_RPS_TARGET`/`PII_APP_PORT`.
- `Dockerfile` — multi-stage `golang:1.27-alpine` → `alpine:3.19`, `CGO_ENABLED=0`.
- `demo/selfcheck.py` — покрытие типов ТЗ, roundtrip 100%, FP (Пушкин/отделение/голая дата/ПИН без карты), регистр.
- `process_api.yaml` — OpenAPI-спека контракта (§5.1).
- `scripts/pack.sh` — zip только исходников (без target/.git/__pycache__/бинарников).
- `README.md` — переписан на 5 предложений (compose, `/process`, `/systems`, env, ПД не логируются).

### Трек 6: закрытие пробелов детекции (2026-09-22)

Selfcheck выявил, что «место рождения» и «гражданство» (типы ТЗ) не маскировались:
- `detect/ner.go` — добавлены `birthPlaceRe` (место рождения/родился/родилась) и `citizenshipRe` (гражданство/гражданин) → типы `birth_place`/`citizenship`.
- `mask/rules.go` — добавлена `MaskWord` (первая буква + `*`), зарегистрированы `birth_place`/`citizenship`.
- Тесты: `ner_test.go` (birth_place/citizenship), `rules_test.go` (MaskWord).
- ПИН без карты — корректно не маскируется (гейт §6), перенесён в FP-кейсы selfcheck.

### Verification

```text
go build ./...   # ok
go test ./...    # ok (api, compute, detect, mask)
go vet ./...     # ok
gofmt -l .       # clean
```

Smoke-test (Redis на :6390): mask → demask roundtrip 100%, `mask_ok == demask_ok`, 422 на пустое тело, `/systems` + `/clear` работают.
`python demo/selfcheck.py` → `SELFCHECK OK` (masked 14/14, roundtrip 14/14, fp 4/4).
`docker compose up --build` — образ собирается (порт 6379 занят локальным redis — env-конфликт, не compose).

## Дальше

1. `demo/load.py` (разгон 330→1000, mean/p50/p95/p99, mask_ok==demask_ok) — Трек 7.
2. Трек 8: финальный `go vet`/`gofmt`, сухой прогон `pack.sh`.

## Блокеры

Нет.