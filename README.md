# Balancer — модуль безопасности ПД

Сервис маскирования/демаскирования персональных данных в цепочке
«система-потребитель → LLM». Контракт проверки: `POST /process`.

Стек: Go (один модуль). `internal/api` (тонкая дверь) + `internal/compute`
(детекция, маска, vault) + Redis. Взаимодействие api ↔ compute — на уровне
пакетов (прямой вызов функций), HTTP-слой один — `cmd/server`.

---

## Карта Markdown-файлов

### Спека и план (читать по задаче)

| Файл | О чём |
|---|---|
| [`main.md`](main.md) | Официальное ТЗ трека: функциональные/нефункциональные требования, контракт `POST /process` (Приложение A), правила нагрузочного теста и эталона (Приложение B). |
| [`требования.md`](требования.md) | Краткая выжимка требований (дубль/черновик формулировок ТЗ). Для агентов и людей удобнее опираться на `main.md`. |
| [`balancer.md`](balancer.md) | Практики Go-шлюза 2026 и LLM-gateway: что строить / не строить, рыночная карта решений по пунктам ТЗ. |
| [`plan.md`](plan.md) | План реализации: инварианты (эталон на forward + vault на reverse), must-порядок, треки агентов, selfcheck/load. |
| [`hack.md`](hack.md) | Подробный playbook дня: архитектура (Go-пакеты api/compute), алгоритмы детекции/маски, промпты треков, критерии приёмки. |

### Харнес агентов (как работать в репо)

| Файл | О чём |
|---|---|
| [`AGENTS.md`](AGENTS.md) | Корень харнеса: workflow, команды проверки, lookup по ключевым словам, границы контракта. Читать первым. |
| [`CLAUDE.md`](CLAUDE.md) | Shim для Claude Code → указывает на `AGENTS.md`, не второй свод правил. |
| [`docs/agents/OWNERSHIP.md`](docs/agents/OWNERSHIP.md) | Exclusive write-globs для параллельных агентов (треки A–E) и шаблон spawn-промпта. |
| [`agent-progress.md`](agent-progress.md) | Индекс прогресса; детали параллельных треков — в `logs/agent-progress-<track>.md`. |
| [`logs/README.md`](logs/README.md) | Назначение папки логов прогресса треков. |

### Cursor rules (не `.md`, но часть харнеса)

| Файл | О чём |
|---|---|
| [`.cursor/rules/agents-harness.mdc`](.cursor/rules/agents-harness.mdc) | Always-on: читать `AGENTS.md`, verify, ownership. |
| [`.cursor/rules/go.mdc`](.cursor/rules/go.mdc) | Конвенции Go при работе в `internal/**/*.go`. |

### Связанный JSON (не md)

| Файл | О чём |
|---|---|
| [`feature_list.json`](feature_list.json) | Статус фич, owner-track, write_globs, команды verification. Пишет glue/человек. |

---

## С чего начать

1. **Человек / жюри:** `main.md` → поднять сервис (когда будет compose) → `POST /process`.
2. **Реализация:** `plan.md` → при необходимости детали в `hack.md` / `balancer.md`.
3. **Агент:** `AGENTS.md` → ownership → код; прогресс в `agent-progress.md`.