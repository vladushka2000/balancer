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

### Трек 7 (нагрузка, 2026-09-22)

- `demo/load.py` — генератор нагрузки под профиль жюри:
  - Профиль `jury`: разгон 50 → avg 330 → пики 1000 → спад (ramp_profile).
  - Профиль `rps2000`: avg 1500 → пики 2000 (плюс ТЗ, не gate).
  - Смешанные типы ПД (паспорт, ФИО, дата, в/у, адрес, email, телефон, ИНН, карта, CVV, ПИН, код подразделения).
  - На каждом payload_id: POST mask → POST demask с НАШИМ result.
  - Отчёт: RPS факт, latency mean/p50/p95/p99, count_429, mask_ok, demask_ok, roundtrip_fail.
  - Gate: mask_ok == demask_ok, roundtrip_fail == 0, p99 ≤ 1с.
- Проверено на mock-сервере: mask_ok == demask_ok, roundtrip_fail == 0, LOAD OK.

### Трек 7: фикс pair invariant (2026-09-22)

- Selfcheck показал `mask_ok=26272 != demask_ok=23536` — нарушение pair invariant.
- Причина: `demo/load.py` выбирал payload_id **случайно** (`random.randrange`), из-за чего
  один pid маскировался несколько раз (ретрай маски → `mask_ok++` без пары demask) и
  demask мог прийти раньше своей маски (lookup miss → сервер считал это новой маской).
  Клиентский отчёт считал по HTTP-статусу (все 200 → равны), а серверный `/stats`
  (читает selfcheck) — по направлению, поэтому разошёлся.
- Фикс: каждый pid генерируется потокобезопасным счётчиком (`PidCounter`) и используется
  ровно один раз для mask + demask, строго по порядку (как требует жюри).
- В отчёт добавлены серверные `server_mask_ok`/`server_demask_ok` из `/stats`; gate теперь
  включает `server_mask == server_demask`.
- Проверено на mock-сервере: `server_mask_ok == server_demask_ok`, LOAD OK.

### Трек 7: фикс pair invariant (roundtrip_fail + server imbalance, 2026-09-22)

- После фикса pid появился `roundtrip_fail: 1704` и `server_mask_ok != server_demask_ok`.
- **Корень 1 (roundtrip_fail)**: pid переиспользовались между прогонами (`load-{n}`),
  store в Redis персистентен → кросс-ран interference. Фикс: `PidCounter` с run-id
  (`load-{timestamp}-{n}`) — pid уникален глобально.
- **Корень 2 (server imbalance)**: `random_email` генерировал кириллические email
  (`пётр.петров@bank.ru`), которые НЕ детектились regex `[a-zA-Z0-9._%+-]+@...`.
  Для таких payload маска == оригинал → demask считался как mask retry
  (`DirectionMask`), раздувая `mask_ok`. Фикс: ASCII email-имена (`EMAIL_NAMES`).
- В отчёт добавлены серверные `server_mask_ok`/`server_demask_ok` из `/stats`;
  gate включает `server_mask == server_demask`. Перед чтением `/stats` — пауза 2.5с
  (stats публикуются в Redis каждые 2с, иначе publish-lag даёт ложный off-by-1).
- `cmd/server/main.go`: Redis pool `PoolSize: 256, MinIdleConns: 16` (дефолт
  `10*GOMAXPROCS` мал под конкуренцией → 500 на mask из-за pool exhaustion).
- Проверено на чистом сервере и на реальном nginx (8080): `LOAD OK`,
  `mask_ok == demask_ok`, `roundtrip_fail == 0`; `selfcheck.py` → `SELFCHECK OK`.

### Трек 7: фикс pair invariant — no-PII payload (server-side, 2026-09-22)

- После фикса email остался `server_mask_ok != server_demask_ok` на FP-кейсах
  selfcheck (payload без ПД: «Пушкин», «адрес отделения»).
- **Корень**: для payload без ПД маска == оригинал (`stored.mask == stored.original`).
  `Store.Lookup` проверял `payload == rec.Original` ПЕРВЫМ → demask такого payload
  считался как mask retry (`DirectionMask`), раздувая `mask_ok`.
- **Фикс**: в `internal/compute/store.go` `Lookup` проверяет `payload == rec.Mask`
  (demask) ДО `payload == rec.Original` (mask retry). Для no-PII payload demask
  теперь корректно считается `DirectionDemask`; для обычного payload поведение
  не меняется (mask != original).
- Проверено на реальном nginx (8080): `LOAD OK` (`mask_ok == demask_ok`,
  `roundtrip_fail == 0`), `selfcheck.py` → `SELFCHECK OK` (fp 4/4, pair invariant ok).

### Трек 7: фикс pair invariant — selfcheck FP-кейсы (2026-09-22)

- После фикса store остался `server_mask_ok != server_demask_ok` на FP-кейсах
  selfcheck («Пушкин», «адрес отделения» — payload без ПД).
- **Корень 1**: `check_masked` и `check_no_pii` делали ОДИН POST (маску) без демаски
  → раздували `mask_ok`. Фикс: каждый кейс прогоняется как mask→demask
  (`mask_and_demask`), проверяя `mask != original` (ПД) / `mask == original` (FP).
- **Корень 2**: фиксированные payload_id (`pii-1`, `fp-1`) переиспользовались между
  прогонами → для no-PII payload (mask == original) повторный mask считался demask
  (неоднозначность). Фикс: уникальные payload_id на прогон (`pii-{run_id}-n`,
  `fp-{run_id}-n`).
- **Корень 3**: `check_pair_invariant` читал `/stats` до публикации (stats в Redis
  публикуются каждые 2с) → ложный off-by-N. Фикс: пауза 2.5с перед чтением `/stats`.
- Проверено на реальном nginx (8080): несколько прогонов `selfcheck.py` → `SELFCHECK OK`,
  `load.py` → `LOAD OK` (`mask_ok == demask_ok`, `roundtrip_fail == 0`).

### Оптимизация CPU-intensive операций (2026-09-22)

По [`optimize.md`](optimize.md), все правки эквивалентны по результату (спаны/маски не меняются):

- `mask/masker.go` — `Apply` O(n²) → O(n): спаны по возрастанию Start, один проход, один `strings.Builder`.
- `mask/rules.go` — `MaskRules()` → package-level `var maskRules` (map только читается).
- `compute/processor.go` — `tokenCount` byte-цикл без аллокаций (убрал `strings`).
- `compute/crypto.go` + `store.go` — AEAD (GCM) кэшируется один раз в `Store`; `encrypt`/`decrypt` принимают готовый AEAD.
- `compute/stats.go` — `RecordTokens` триммит `window` старше 1с (нет memory leak).
- `detect/ner.go` — 1 chunk синхронно; много chunks — bounded worker pool (`min(len, NumCPU)`).
- `detect/context.go` — `windowOf` считает `ToLower` один раз; `hasStructural`/`hasOrgMarker` принимают window.
- `compute/processor.go` — Redis `AliveCount` вынесен из-под `bucketMu` (снижен контеншн).
- `detect/structural.go` — `cardRe` без вложенного квантора; `strings.Map` → `extractDigits` byte-цикл.

Verification: `go build ./...`, `go vet ./...`, `gofmt -l .` (clean), `go test ./...` — все зелёные.

### Трек 8 (качество кода + zip, 2026-09-22)

- `gofmt -l .` — clean; `go vet ./...` — чистый; `go build ./...` — ok; `go test ./...` — все зелёные.
- Нет `panic("todo")`, `log.Fatal`/`os.Exit` на горячем пути `/process` (только на старте сервера).
- Нет deprecated API: redis `go-redis/v9 v9.22.0` (актуальный мажор), не используются `UnstableResp3`/`RawResult`/`RawVal`/`DisableIdentity`.
- Нет закомментированного кода; только пакетные `//`-доки (допустимы). `redact`-режим маскера — только в тестах, не на горячем пути (main использует `partial`).
- DRY: детекторы используют общий `Detector` interface + `Registry` в `base.go`; валидаторы (Luhn/ИНН/СНИЛС/дата) вынесены в хелперы; regex — package-level vars.
- `scripts/pack.sh` — добавлены исключения `ds.pdf`, `ds.md`, `demo/__pycache__/` (датасет/бинарник/кэш не должны попадать в zip).
- Сухой прогон: распакованный zip собирается (`go build`), проходит `go vet`, все тесты зелёные; размер 117K, без `target/`, `.git`, бинарников, датасетов.
- `README.md` — 1 предложение (≤5).

### Трек 8: фикс замечаний аудита (2026-09-23)

- `cmd/server/main.go` — удалён мёртвый конфиг `rpsTarget` и чтение `PII_RPS_TARGET`
  (поле не использовалось; реальный bucket строится из `PII_GLOBAL_RPS`).
- `docker-compose.yml` — добавлен `PII_STORE_KEY` во все 3 server-сервиса
  (иначе AES-GCM шифрование original/mask в Redis отключалось).
- Проверено: `go build`/`go vet`/`gofmt` clean, `go test ./...` зелёные,
  `docker compose config` валиден.

## Дальше

1. Финальный прогон `docker compose up --build` + `selfcheck.py` + `load.py` перед сдачей.

## Блокеры

Нет.