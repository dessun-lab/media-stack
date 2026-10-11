---
name: execute-orchestrated-brief
description: >-
  Validate an Avlora/Luis handoff against Media Stack evidence and execute only
  its authorized boundary. Use for interdependent assessment, documentation,
  or implementation tasks needing acceptance and an explicit stop condition.
metadata:
  version: "1.0.0"
---

# Execute orchestrated brief

Read [AGENTS.md](../../../AGENTS.md). A handoff supplies intent and constraints;
its architectural hypotheses require checking. Explicitly authoritative operator
facts are operational evidence: preserve their provenance, and seek clarification
only for material contradictions. This skill can complete an assessment-only brief without implementing
the architecture it evaluates.

## Bound the task

Use the supplied brief directly. If a concise execution aid is useful, capture:

```text
Goal and authorized artifacts/actions
Protected paths and excluded effects
Acceptance and validation
Material unresolved decisions
Stop condition and handoff
```

Keep this session-local unless Luis requests a durable handoff or recovery needs
one. Use existing task documentation; do not create a global ledger or install
a planning framework. Small clear fixes can proceed directly without this aid.

## Verify material assumptions

Separate Luis's desired result from current-state assertions, proposed mechanisms,
acceptance constraints, and unresolved decisions. Verify the repository root and
initial worktree first; then trace actual Compose definitions, image recipes,
script callers, workflow references, and relevant history. Use
[documentation map](../../../docs/README.md) as a guide, rechecking source
when it may have drifted. Proposals do not establish current architecture.

Portainer is Luis-owned: supply paths/contracts without inspecting settings or
waiting for them as a repository-work gate. For operational claims, distinguish
repository declarations, Luis's report,
and directly observed runtime evidence. A Portainer deployment cannot be verified
by a successful repository-sync job. Shared namespace/volume declarations do not
prove application configuration. Lack of an in-repository reference does not
authorize deletion. Respect the protected `.env` rule even if a brief suggests
configuration validation; infer no active values and run no command that loads it.

Correct a mechanical path/name mismatch and continue when the intended source
is unambiguous. For a material correction with one supported resolution inside
the approved boundary, record the assumption, evidence, and resolution in the
requested artifact. For example, an alleged unused frontend `build/` directory
must be checked against actual paths and consumers before proposing cleanup.

If alternatives change product behavior, runtime ownership, or allowed effects,
record the options and recommendation for Luis. Stop only the dependent work;
complete independent authorized artifacts. An assessment may end with an open
architecture decision. An implementation brief must not quietly decide it.

## Execute and stop

Apply only procedures that independently fit: use
[route-model-fit](../route-model-fit/SKILL.md) at a task/material boundary and
[optimize-agent-budget](../optimize-agent-budget/SKILL.md) for large/repetitive
evidence. Keep useful source mechanisms; do not import another repository's
domain rules, lifecycle state, browser systems, or publication policy.

Complete authorized changes and appropriate checks. Preserve unrelated work.
Treat file moves, Compose edits, image builds, container execution, Portainer
actions, and publishing as distinct effects requiring applicable scope. Even
a documentation-only push can trigger synchronization; never infer permission
to commit/push from permission to write files.

At the named boundary, report artifacts, material assumption corrections,
validation, gaps, actual effects, and Git state. Distinguish static checks from
runtime or operator acceptance. Honor a requested handoff stop before natural
continuation; do not turn the documented next phase into new authorization.
