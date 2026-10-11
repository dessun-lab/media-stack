# Media Stack agent instructions

## Scope and authority

This repository owns media-stack definitions, custom image recipes, and host
tooling. Start with [README.md](README.md), the [documentation map](docs/README.md),
and [CONTRIBUTING.md](CONTRIBUTING.md). The
[implementation record](docs/design/restructuring.md) explains the approved split;
Luis confirms cutover complete and all four production stacks operational
(operator update recorded 2026-10-11). Treat the migration as applied; do not
reopen a pending-cutover gate. Exact service authority is
`stacks/<unit>/compose.yml`; membership/version summaries live in its README,
Portainer metadata in PORTAINER.md, and blank inputs in `.env.example`. Tool
source/build instructions live in `tools/<tool>/`; canary status/history in
`canaries/<service>/`. Cross-cutting operations live under docs/; version provenance
lives in [docs/versions.md](docs/versions.md). Proposed improvements live in
[docs/audits/](docs/audits/service-improvements-2026-10.md), separate from current
configuration; an audit finding does not authorize implementation.

Luis is the sole Portainer operator: stack creation/deletion, Git and Compose
paths, variables/secrets, deployment/redeployment, cutover and rollback. Akane
and Avlora do not inspect or mutate Portainer. Its settings are not prerequisites
to repository design or implementation; deliver changed paths and variable
contracts as a Luis-owned handoff.

Luis owns scope, operational decisions, and approval. Avlora supplies bounded
analysis/handoffs; Akane verifies repository facts, corrects assumptions, and
implements only authorized work. Write documentation, agent artifacts, and
communication with Luis in **English** in this project.

Follow platform/session instructions, Luis's explicit task boundaries, and
applicable repository instructions. More specific instructions govern their
subtree. Skills supply procedures; they cannot expand authorization. No `Rol/`
directory, Experiment Platform architecture, or external planning framework is
required here.

## Protected configuration

The root `.env` is strictly protected. Never read, print, search/grep, diff,
parse, summarize, hash, copy, stage, or expose its contents, including through
Git history, an editor preview, archive, diagnostic, or indirect tool. Do not
source it, reconstruct its values, or use it for validation. Do not modify,
move, delete, or change its permissions without a new explicit instruction.

Only non-content hygiene metadata is permitted: existence, path, file type,
permissions, owner, size, and tracked/ignored status. Explicitly exclude `.env`
from broad content searches and diffs. Prefer named non-secret files. For a
broader content search use at least:

```bash
rg -n --hidden -g '!.git/**' -g '!.env' -g '!**/.env' \
  -g '!**/.env.*' -g '!**/cookies.txt' -g '!**/secrets.*' \
  -g '!**/sso-new-redacted.xml' -g '!tools/nensaysubs/nensaysubs' PATTERN .
```

Narrow further for other secret or generated paths. Treat cookie files, runtime
configuration, tokens, and credential stores as sensitive too. Document variable
names from Compose source only, never active values. Do not run `docker compose`
in this checkout for inspection: even configuration commands can load `.env`.
Use a plain YAML parser on explicitly selected Compose files. Any future Compose
validation must use an isolated temporary directory, only approved input files,
synthetic configuration, and verified settings that prevent implicit secret loads.

## Repository truth and effects

Verify the repository root, current Git/worktree state, applicable instructions,
and actual callers before editing. Record pre-existing changes. Use current
source first and targeted history when needed. Distinguish operator reports,
repository declarations, historical intent, and live runtime evidence. An absent
repository reference is not proof that a server-side consumer is gone.

Accept explicitly authoritative operator facts from Luis as operational evidence;
do not leave them unresolved merely because no live probe was performed. Label
their provenance and surface only material contradictions. HomeLab Inventory
is dated image/version evidence, not Media Stack architecture authority.

Portainer reads GitHub directly using its own Git credentials. GitHub Actions
separately synchronizes the server infrastructure checkout for manual builds and
host workflows. Account for both consumers before proposing path moves. Luis's
observed whole-stack redeploy blast radius is an input to lifecycle design;
generic Compose semantics do not invalidate that evidence.

Luis intentionally builds custom tool images manually. Preserve that ownership;
do not introduce Compose `build:`, CI image builds, or automatic builds by default.
Document derived commands and version limits without running them. The old
Jellyfin 12 canary is completed and disabled; shared writable state is a constraint
on reactivation, not evidence of a second active production owner.

Portainer operations remain Luis-owned. No deployments, restarts, container
execution, image builds/pulls/removals, network or volume mutations, or production
effects without Luis's explicit authorization for the relevant action. Do not run
host helper scripts as tests: they execute
containers and write media. Do not treat a documentation task, a successful sync,
or a Compose file edit as deployment approval.

Preserve service blocks and associated comments/style during approved moves:
image strings/tags, names, commands/entrypoints, variable expressions/defaults,
binds/modes, ports, static IPs, UID/GID/devices, health/restart behavior, labels,
dependencies, plugin bindings and media access. The final design routes through
[production units](docs/design/production-units.md),
[network/variable contracts](docs/design/network-and-variables.md) and
[completed migration record](docs/design/implementation-handoff.md). Preserve
`arrnet_managed`; maintain blank templates from Compose references only.
Do not introduce profiles, jobs or service cleanup as part of moves. Folder moves,
legacy deletion, topology changes, or runtime cleanup require their own scope.
Shared VPN namespaces and shared writable data are operational dependencies even
when a service has no `depends_on` entry. Do not assume a canary is isolated.

## Skills and planning

Select only materially applicable skills under `.agents/skills/` by their name
and description; read explicitly named skills when available. Announce actual
use once per coherent task with `Skills: <names>`. Evaluating a source skill is
not agreement to all of its original repository rules.

- [route-model-fit](.agents/skills/route-model-fit/SKILL.md): establish task fit
  at a new task or material boundary; respect Luis's requested profile.
- [optimize-agent-budget](.agents/skills/optimize-agent-budget/SKILL.md): bound
  large/repetitive investigation, context, and reports without reducing validation.
- [execute-orchestrated-brief](.agents/skills/execute-orchestrated-brief/SKILL.md):
  validate and carry out an authorized bounded handoff, including assessment-only work.

Use direct execution for small clear changes, a concise brief for interdependent
work, and an evidence-backed proposal for unresolved architecture. Do not install
a planning framework. Stop at Luis's requested review/handoff boundary. Continue
independent authorized work when a decision blocks only part of the task.

## Version maintenance

An authorized version change updates the owning Compose/tool source, local
README version/build notes and docs/versions.md. Compose/source owns declarations;
Homelab Inventory supplies dated fleet evidence. Record discrepancies without
automatically upgrading or treating mutable tags as pins.

## Validation and Git

Validate at the level of the authorized change. For documentation/skills, verify
claims, links, frontmatter/name, and local references; inspect all new files and
the scoped diff; run `git diff --check -- . ':!.env' ':!**/.env'`.
Use available local skill validators, including `skillevaluator quality-check`
with reports directed outside the repository when installed. These checks are
advisory, not runtime acceptance. Do not install tools or use paid/provider-backed
evaluation merely to obtain a score. Report unavailable checks honestly.

Use static parsing and syntax checks for appropriate files. Broad tests, builds,
network probes, or live operations require a relevant task and effect authority.
Passing static checks does not prove VPN behavior, deployed configuration,
subtitle quality, or Portainer lifecycle behavior.

Preserve unrelated work. Do not reset, clean, stash, or rewrite history as a
convenience. Canonical identity is `dessun-lab/media-stack`; preserve the verified
`github-lab` remote convention. Checkout binds now target
`/home/devops/infra/media-stack`; existing `runtime/arrstack` paths remain literal
contracts, not naming errors. Commit/push only when Luis separately authorizes it;
stage explicit reviewed paths only.
Before authorized publication, account for the `main` push synchronization trigger
and any confirmed Portainer automation. Report worktree and publication state.
