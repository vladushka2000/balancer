# Прогресс

Индекс. Параллельные треки пишут в `logs/agent-progress-<track>.md`.
`feature_list.json` обновляет только glue/человек.

## Текущий фокус

Харнес и ресёрч-артефакты готовы. Код реализации — по [`plan.md`](plan.md).

## Сделано (недавнее)

### Docs + harness (2026-09-22)

- [`balancer.md`](balancer.md) — практики Go-шлюза 2026, LLM-steal list, рыночная карта `main.md`.
- [`plan.md`](plan.md) — must-порядок, инварианты (etalon forward + vault reverse), треки.
- Корневой [`AGENTS.md`](AGENTS.md), shim [`CLAUDE.md`](CLAUDE.md), `.cursor/rules/*`.
- [`feature_list.json`](feature_list.json), [`docs/agents/OWNERSHIP.md`](docs/agents/OWNERSHIP.md).

### Verification

```text
# docs-only change; no go/cargo run required for this step
ls AGENTS.md CLAUDE.md balancer.md plan.md feature_list.json docs/agents/OWNERSHIP.md
```

## Дальше

1. Track A: каркас `api/` (Go).
2. Track B/C/D: detect / mask / store в `compute/` (worktrees).
3. Glue: pipeline, compose, selfcheck, load, zip.

## Блокеры

Нет.
