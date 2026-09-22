# AGENTS.md — compute/

Rust engine for PII module. Parent rules: root [`AGENTS.md`](../AGENTS.md).

## Role

Detect → context → merge → etalon partial mask → vault (`payload_id`).
Admin `/systems`, `/stats`. Demask = store lookup (no detect semaphore).

## Module map

| Path | Responsibility |
|---|---|
| `src/detect/structural.rs` | Regex + checksums |
| `src/detect/ner.rs` | Dict/regex FIO, address, orgs |
| `src/detect/context.rs` | Pushkin / bank branch gate |
| `src/detect/merge.rs` | Overlaps + PIN gate |
| `src/mask/` | Partial rules, masker |
| `src/store.rs`, `keys.rs` | Correspondence vault |
| `src/pipeline.rs` | Orchestration |
| `src/routers/process.rs` | `/process` |
| `src/routers/admin.rs` | `/systems` `/stats` `/health` `/clear` |
| `src/stats.rs` | mean/p50/p95/p99, mask_ok/demask_ok |
| `src/ratelimit.rs` | Detect concurrency semaphore |

## Commands

```bash
cd compute && cargo test
cd compute && cargo clippy -- -D warnings && cargo fmt --check
```

## Boundaries

- Respect track globs (B detect / C mask / D store / E glue). See `docs/agents/OWNERSHIP.md`.
- Put Redis **before** 200 on new mask. No PII in logs.
- No comments in code. No `todo!()` on `/process` path in release.
- FPE/synthetic — after must features green.
