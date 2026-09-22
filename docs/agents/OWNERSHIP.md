# Ownership — parallel agents

Exclusive **write** globs. GitHub CODEOWNERS is for human review; this table is
for agents: **one writer per path**. Parallel = **git worktree per agent**.

## Table

| Track | Owner | Write globs | Read-only OK | Never touch |
|---|---|---|---|---|
| A | go-api | `api/**` | `compute/src/models.rs` (contract shape) | `compute/**`, `docker-compose.yml`, `feature_list.json` |
| B | rust-detect | `compute/src/detect/**`, `compute/tests/test_structural.rs`, `test_ner.rs`, `test_context.rs`, `test_merge.rs` | `compute/src/models.rs`, `pipeline.rs` (signatures) | `api/**`, `compute/src/mask/**`, `store.rs` |
| C | rust-mask | `compute/src/mask/**`, `compute/src/dicts/**`, `compute/tests/test_mask*.rs` | detect `Span` type | `api/**`, `compute/src/detect/**` |
| D | rust-store | `compute/src/store.rs`, `keys.rs`, encrypt helpers, `compute/tests/test_store.rs` | `models.rs` CorrRecord | `api/**`, `detect/**`, `mask/**` |
| E | glue | `docker-compose.yml`, `compute/src/pipeline.rs`, `routers/**`, `stats.rs`, `main.rs`, `demo/**`, `scripts/**`, `README.md`, `feature_list.json`, DTO sync | everything | starts **after** A–D PRs |

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
