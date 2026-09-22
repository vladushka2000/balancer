# Plan — реализация PII-модуля

Источник истины для кода. Уже: [`main.md`](main.md) (контракт),
[`balancer.md`](balancer.md) (практики), [`hack.md`](hack.md) (детали дня).
Харнес: [`AGENTS.md`](AGENTS.md), ownership: [`docs/agents/OWNERSHIP.md`](docs/agents/OWNERSHIP.md).

**Стек (зафиксировано):** весь runtime на **Go**. Rust `compute/` снят и
переписан. Один Go-модуль, два пакета: `internal/api` (дверь) + `internal/compute`
(движок), вызовы функций без HTTP. SSE-шлюз (`gateway`) **не нужен** — контракт
синхронный, чекер бьёт один `POST /process`.

---

## 0. Цель и инварианты

Сервис: `POST /process` `{payload, payload_id} → {result}`.
Первый запрос с id — маскирование; второй с нашей маской — демаскирование.

| Инвариант | Правило |
|---|---|
| **Etalon forward** | Маска близка к эталону (норм. span-Levenshtein). Стиль: `И. И. И.`, `45** ****56`, PCI BIN+last4 |
| **Vault reverse** | Demask: exact `payload == stored.mask` → `original`. Не инвертировать `*` |
| **Pair** | `mask_ok == demask_ok` на прогоне |
| **Sync put** | Redis/write-through **до** HTTP 200 на новой маске |
| **No 429 on demask** | Lookup/retry **без** семафора детекции |
| **No PII in telemetry** | Логи/метрики: types, offsets, lengths, `payload_id`, timings |
| **Contract frozen** | Тела `ProcessRequest` / `ProcessResponse` не менять |
| **Zip gate** | Только исходники; `gofmt` + `go vet` чистые |
| **Go-only runtime** | `internal/api` + `internal/compute` — Go. Без Rust на hot path |

---

## 1. Архитектура файлов

```
cmd/server/               # HTTP: POST /process, /app/health, admin
internal/api/             # Go door: validate, rate limit, вызов compute
internal/compute/         # Go engine: detect, mask, store, pipeline, stats
internal/models/          # Замороженный контракт (DTO)
docker-compose.yml        # redis + server
demo/{selfcheck,load}.py
scripts/pack.sh
README.md                 # ≤5 предложений
```

api ↔ compute — **на уровне пакетов** (прямой вызов функций), не HTTP.

```mermaid
flowchart LR
  Checker["AlfaSonar"] -->|"POST /process"| Server["cmd/server"]
  Server -->|"api.Door.Process"| Api["internal/api"]
  Api -->|"compute.Processor.Process"| Compute["internal/compute"]
  Compute --> Detect["regex checksum context"]
  Compute --> Store["Redis vault"]
  Detect --> Mask["etalon partial"]
  Mask --> Store
  Store -->|"put then 200"| Server
```

---

## 2. Must-порядок

| Шаг | Что | Done when |
|---|---|---|
| M0 | Решение Go-only: каркас `internal/api` + `internal/compute`; Rust не на hot path | `go build ./...` |
| M1 | Контракт `/process` + health | curl 200; 422 на пустое тело |
| M2 | Типы ПД из ТЗ + регистр + «серия/номер» + checksums | unit tests green |
| M3 | Etalon-style partial masks | selfcheck Levenshtein ≤ порога на фикстурах |
| M4 | Vault + roundtrip 100%; put до 200 | demask == original; pair counters |
| M5 | FP: Пушкин / отделение; дата только с маркером; PIN iff card | context tests |
| M6 | Latency mean/p50/p95/p99 на разгоне 330→1000; limiter ~1500 | load report; 429≈0 на пике |
| M7 | Логи без ПД; README ≤5; zip + gofmt/vet | pack.sh dry-run; `go vet ./...` |

**Plus (после must):** FPE FF1, synthetic, RPS 2000, другие УЛ, admin UX.

---

## 3. Треки и exclusive globs

Параллель = **git worktree на агента**. Shared (`compose`, DTO, `feature_list.json`) — только glue.

| Track | Owner | Write | Never |
|---|---|---|---|
| A | go-api | `internal/api/**`, `cmd/server/**` | `internal/compute/**` |
| B | go-detect | `internal/compute/detect/**`, related tests | `internal/api/**`, `mask/**`, `store.go` |
| C | go-mask | `internal/compute/mask/**`, `dicts/**` | `internal/api/**`, `detect/**` |
| D | go-store | `store.go`, `keys.go`, `crypto.go`, redis encrypt glue | `internal/api/**`, `detect/**` |
| E | glue | compose, pipeline wiring, `/stats`, zip, README, selfcheck/load | starts after A–D |

Подробности: [`docs/agents/OWNERSHIP.md`](docs/agents/OWNERSHIP.md).

---

## 4. Go api (track A)

1. `POST /process`: decode → validate non-empty → rate limit → compute.
2. Token bucket default **1500**; `429` + `Retry-After: 1`.
3. Вызов `compute.Processor.Process` — прямой вызов функции, без HTTP.
4. Map errors: validation → 422, rate limit → 429, compute → 500.
5. `GET /app/health` → 200.
6. Logs: method, path, status, duration, `payload_id` — **не** payload.
7. Tests: 200/422/429; no payload in logs.

Стек: Go 1.22+, stdlib `net/http`, `log/slog`. Без Redis в api.

---

## 5. Go compute

### 5.1 Detect (track B)

- Structural: passport, driver license, INN (+checksum), phone, email, card+Luhn, CVV, PIN, postal, department code, birth/issue date **with marker**.
- NER/dict: FIO, address components, birth place, issuing org, citizenship, cardholder.
- Context: window + markers; famous.txt / org_addresses.txt.
- Merge: prefer regex; PIN without card → drop.
- Engine: `regexp` case-insensitive; compile at startup; chunks for NER (~4k/200).

### 5.2 Mask (track C)

- Per-type **partial** in etalon style (таблица в `hack.md` §4.2).
- Apply spans right-to-left.
- Modes `fpe`/`synthetic`/`redact` — stubs only until must green.

### 5.3 Store + process (track D + glue)

- `lookup(payload_id, payload)`: original→mask | mask→original | None.
- Hit → return immediately (**no** detect semaphore).
- Miss → acquire semaphore → pipeline → **sync** `put` (AES-GCM) → 200.
- Redis put fail → **500**, not 200.
- TTL default **86400**.
- Stats: mean/p50/p95/p99, RPS, TPS est, `mask_ok`, `demask_ok`, `count_429`.

### 5.4 Admin (glue, off checker path)

`POST/GET /systems`, `GET /stats`, `GET /health`, `POST /clear`.
`/process` uses default-policy (`types=all`, `partial`, demask on).

### 5.5 Миграция с Rust

1. Rust `compute/` удалён; вся логика переписана на Go в `internal/compute`.
2. Новые фичи — сразу в Go; Rust-дерева в репо нет.
3. Zip **не** включает `target/`, `.rs` — только Go-исходники.

---

## 6. Selfcheck и нагрузка

### `demo/selfcheck.py`

Gate:

1. Покрытие типов ТЗ (маска закрывает ПД).
2. Forward: расстояние до эталонных строк (стиль ТЗ / Приложение B).
3. Reverse: demask(наш result) == original; `mask_ok == demask_ok`.
4. FP: Пушкин, адрес отделения — не маскируются.
5. Голая дата — нет; «дата рождения …» — да.
6. Регистр: `ПАСПОРТ 4509 123456` находится.

Формат контракта: `https://process-test.holydev.space/process` (только формат обмена).

### `demo/load.py`

Профиль `jury`: разгон 50 → ~330 → пики 1000.
На каждый id: mask, затем demask **нашей** маской.
Отчёт: RPS, mean/p50/p95/p99, 429, mask_ok, demask_ok, roundtrip fails.

### `scripts/pack.sh`

Zip исходников без `target/`, `.git`, `venv`, бинарников, датасетов.

---

## 7. Env (минимум)

| Var | Default | Где |
|---|---|---|
| `PII_RPS_TARGET` | 1500 | api |
| `REDIS_URL` | `redis://localhost:6379/0` | compute |
| `PII_CORR_TTL_SEC` | 86400 | compute |
| `PII_MAX_CONCURRENT` | num_cpus | compute |
| `PII_SEM_WAIT_SEC` | 0.3 | compute |
| `PII_STORE_KEY` | — | compute (AES-GCM) |

---

## 8. Критерии сдачи

- [ ] Весь hot path на **Go** (`internal/api` + `internal/compute`); Rust не в compose сдачи.
- [ ] `POST /process` по контракту; идемпотентен.
- [ ] Forward: маски в стиле эталона; selfcheck Levenshtein ок.
- [ ] Reverse: 100% original; `mask_ok == demask_ok`.
- [ ] Типы ТЗ + регистр + вариации написания.
- [ ] FP + даты с маркером + PIN-гейт.
- [ ] Latency mean/p50/p95/p99 на разгоне; 429≈0 на пике 1000.
- [ ] Demask не 429; put до 200.
- [ ] ПД не в логах; store encrypted.
- [ ] Zip + `go vet`; README ≤5 предложений.
- [ ] `docker compose up --build` поднимает redis+server.

Статус фич — [`feature_list.json`](feature_list.json). Прогресс — [`agent-progress.md`](agent-progress.md).
