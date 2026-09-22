# CLAUDE.md

Before any action, read and follow [`AGENTS.md`](AGENTS.md). That file is the
single source of truth for workflow, docs lookup, verification, and completion.

## Behavior

- Work only inside the requested scope and ownership glob.
- Small steps; verify each meaningful change (observed output).
- Do not claim completion before checks pass.
- Ask concise questions when requirements are unclear.

## Safety

- No unrelated refactors or dependency churn.
- No destructive ops without confirmation.
- Do not edit `feature_list.json` / compose / frozen DTOs unless you are glue.

## Progress

- Multi-step: update [`agent-progress.md`](agent-progress.md) or
  `logs/agent-progress-<track>.md`.
- Keep [`feature_list.json`](feature_list.json) aligned (glue-owned).
