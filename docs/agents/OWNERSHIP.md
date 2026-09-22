# Ownership — parallel agents

Exclusive **write** globs. GitHub CODEOWNERS is for human review; this table is
for agents: **one writer per path**. Parallel = **git worktree per agent**.

## Table

| Track | Owner | Write globs | Read-only OK | Never touch |
|---|---|---|---|---|
| A | go-api | `internal/api/**`, `cmd/server/**` | `internal/models/models.go` (contract shape) | `internal/compute/**`, `docker-compose.yml`, `feature_list.json` |
| B | go-detect | `internal/compute/detect/**`, related tests | `internal/models/models.go`, `pipeline.go` (signatures) | `internal/api/**`, `internal/compute/mask/**`, `store.go` |
| C | go-mask | `internal/compute/mask/**`, `internal/compute/detect/dicts/**`, related tests | detect `Span` type | `internal/api/**`, `internal/compute/detect/**` |
| D | go-store | `internal/compute/store.go`, `keys.go`, `crypto.go`, encrypt helpers, related tests | `models.go` CorrRecord | `internal/api/**`, `detect/**`, `mask/**` |
| E | glue | `docker-compose.yml`, `internal/compute/pipeline.go`, `processor.go`, `stats.go`, `cmd/server/main.go`, `demo/**`, `scripts/**`, `README.md`, `feature_list.json`, DTO sync | everything | starts **after** A–D PRs |

## Rules

1. Disjoint write globs. If two tracks match the same path — not parallel.
2. Shared files are a **queue** (glue or human), not a shared glob.
3. Spawn prompt must repeat: «Edit only `<globs>`. If you need another path, stop.»
4. Do not let four agents rewrite `feature_list.json` or root `AGENTS.md`.
5. Merge via PRs; glue last.
6. Progress: `logs/agent-progress-<track>.md` (create `logs/` as needed).

## Spawn prompt template

```
Track: <A|B|C|D|E>
Write only: <globs from table>
Never: <never list>
Done when: <verification commands from feature_list.json>
Do not edit AGENTS.md / feature_list.json / compose unless you are glue.
Open a PR; do not merge to main.
```