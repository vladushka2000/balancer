# AGENTS.md

## Purpose

Модуль безопасности ПД (PII-guard): маскирование/демаскирование в цепочке
consumer → LLM. Контракт проверки — `POST /process`.
Харнес держит агентов в рамках сессий и параллельных треков.

## Split

Один Go-модуль. api ↔ compute — **на уровне пакетов** (прямой вызов функций),
не HTTP. HTTP-слой один — `cmd/server`.

| Path | Lang | Role |
|---|---|---|
| `cmd/server` | Go | HTTP: `POST /process`, `/app/health`, admin |
| `internal/api` | Go | Thin door: validate, rate limit, вызов compute |
| `internal/compute` | Go | Detect, mask, vault, admin, stats |
| `internal/models` | Go | Замороженный контракт (DTO) |

Контракт заморожен: `{payload, payload_id}` → `{result}`.

## Workflow

1. Прочитай этот файл.
2. Кратко переформулируй задачу.
3. Lookup ниже → нужный файл/док. Подтверди символы в коде.
4. Короткий план → минимальный diff.
5. Прогони проверки; **видел** вывод.
6. Многошаговая работа: обнови `agent-progress.md` / track-лог и
   `feature_list.json` (пишет glue/человек, не четыре агента сразу).
7. Только потом — done.

## Editing

- Маленькие scoped diffs. Без лишних рефакторов и новых deps без запроса.
- **Без комментариев** в Go (gate хакатона).
- Не меняй `ProcessRequest` / `ProcessResponse` без явного решения.
- Параллель: exclusive globs — [`docs/agents/OWNERSHIP.md`](docs/agents/OWNERSHIP.md).
  Один агент = один git worktree. Shared files (compose, DTO, feature_list) — glue.

## Verification

Считается только **наблюдаемый** вывод команды.

| Area | Command |
|---|---|
| Go build | `go build ./...` |
| Go tests | `go test ./...` |
| Go vet | `go vet ./...` |
| Go fmt | `gofmt -l .` |
| Detect slice | `go test ./internal/compute/detect/...` |
| Mask slice | `go test ./internal/compute/mask/...` |
| Store slice | `go test ./internal/compute/ -run TestProcessor` |
| Compose | `docker compose up --build` |
| Selfcheck | `python demo/selfcheck.py --url http://localhost:8080` |
| Load jury | `python demo/load.py --url http://localhost:8080 --profile jury` |
| Zip | `scripts/pack.sh` |

Качество: complexity / DRY / KISS / no deprecated / no dirty code — gate zip (~6к правил).

## Completion / stop

Done: изменение есть, checks green (seen), progress updated if multi-step, no blockers.

Stop and ask: scope unclear; architecture conflict; destructive op; cannot verify safely.

## Docs (pointers — do not paste into always-on)

| Doc | When |
|---|---|
| [`main.md`](main.md) | Контракт, эталон forward, нагрузка чекера |
| [`balancer.md`](balancer.md) | Практики Go-шлюза + рыночная карта |
| [`plan.md`](plan.md) | Must-порядок, треки, инварианты |
| [`hack.md`](hack.md) | Детальный playbook дня |
| Nested `internal/api`, `internal/compute` | Работа внутри пакета |

## Jury (2026-09-22)

- Разгон, avg ~330 RPS, пик 1000; 429 не ошибка, но в статистике.
- Смотрят mean/p50/p95/p99; `mask_ok == demask_ok`.
- Reverse всегда с **нашей** маской; put до 200; demask без 429.
- Forward: сравнение с эталоном (Levenshtein). Подробности — `main.md` / `plan.md`.

## Key concepts lookup

| Keywords | Read / edit |
|---|---|
| contract, /process, payload_id | `main.md`; `cmd/server/main.go`; `internal/models/models.go` |
| detector, Luhn, INN, passport | `internal/compute/detect/structural.go` |
| FIO, address, Pushkin, context | `internal/compute/detect/ner.go`, `context.go` |
| merge, PIN gate | `internal/compute/detect/merge.go` |
| mask, etalon, partial | `internal/compute/mask/` |
| vault, Redis, demask | `internal/compute/store.go`, `keys.go` |
| rate limit, 429 | `internal/api/ratelimit.go` |
| concurrency semaphore | `internal/compute/ratelimit.go` |
| /stats, latency percentiles | `internal/compute/stats.go` |
| /systems, SystemConfig | `internal/compute/repo.go` |
| pipeline | `internal/compute/pipeline.go` |
| compose, deploy | `docker-compose.yml` |
| ownership, parallel tracks | `docs/agents/OWNERSHIP.md` |

## Rules

1. Код — истина при drift с промптом; скажи об этом.
2. ПД не логировать и не класть в метрики.
3. Качество кода — Code quality в nested AGENTS + `go vet`/`gofmt` перед сдачей.
