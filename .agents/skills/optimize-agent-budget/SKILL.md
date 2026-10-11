---
name: optimize-agent-budget
description: >-
  Bound context, investigation, and reporting for large Media Stack audits,
  repetitive work, or explicit budget requests while preserving evidence,
  operational boundaries, and required validation.
metadata:
  version: "1.0.0"
---

# Optimize agent budget

Use this skill when a task is large/repetitive, has an explicit budget, or is
replaying settled context. It is not a requirement to compress every exchange.
The [root instructions](../../../AGENTS.md) govern protected files and effects.

## Bound investigation

Start with exact authorized paths, current Git state, and the acceptance
questions. Prefer a path inventory, then `rg` over selected text, then targeted
source/history reads. Explicitly exclude root `.env`, other secrets, `.git`,
ignored SSO artifacts, and native binaries from broad content searches.
Inspect generated assets only as needed to answer their ownership/reference
question; start with metadata, entrypoint, version, and callers.

Use plain parsing of named non-secret Compose files to summarize services,
dependencies, mounts, and network declarations. Do not invoke Compose in the
checkout for a shorter report: it may implicitly read the protected `.env`.
Never trade secret exclusion for search completeness or tool convenience.

Batch independent reads with bounded output. Preserve enough source paths,
symbols/line references, and commit IDs for meaningful claims to be checked.
Stop expanding a search when evidence answers the question. If a remaining
answer lives in external server configuration, record the compatibility handoff.
Portainer settings are Luis-owned, not an agent inspection target or a blocker
for repository design/implementation.

## Preserve validation and effects

Use static checks for documentation and skill work. Do not build images,
execute helper scripts, query containers, call subtitle sites, or redeploy to
demonstrate diligence. A previous static result can be reused when its inputs
are unchanged; repeat it when new edits or failures justify it.

Prefer installed deterministic tools and standard libraries. No search indexes,
browser controller, analyzer installation, or provider evaluation is required.
If a necessary validator is absent, find an existing local equivalent or report
the unavailable check without claiming success. Do not omit failed checks,
data-integrity concerns, unknown runtime parity, or Luis's required decisions
to meet a reporting budget.

## Communicate deltas

Put detailed topology, evidence, and alternatives in the requested repository
documents. Send progress for material findings, changed conclusions, blockers,
decisions, or meaningful long validation. Respect platform progress requirements;
avoid narrating individual reads and routine successful commands.

Use the final sections Luis requested. Otherwise summarize status, material
artifacts, validation, exceptions, effects/Git state, and one bounded next step.
For example, report three passing skill checks together and separately identify
a Luis-owned runtime handoff; do not repeat three identical success sections
or confuse operator acceptance with static validation.

Carry only current authority, settled decisions, exact unfinished work, evidence
paths, acceptance, and stop conditions into a continuation. Never optimize away
protected-file handling, unrelated user changes, rollback requirements, effect
approval, or substantive validation. Explicit exhaustive-report requests take
precedence over this compact default. Do not estimate quota/cost without data.
