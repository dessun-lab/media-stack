# Contributing to Media Stack

Start with [documentation](docs/README.md) and [AGENTS](AGENTS.md). Write project
documentation and communication in English. Luis owns product/operational choices;
Portainer is exclusively his surface and is never an agent inspection target.
The four-stack cutover is complete per Luis's operator update recorded 2026-10-11.

## Choose the owner

- Production services join one of the four `stacks/<unit>/` operational units;
  preserve hard VPN namespace/dependency ownership. Compose owns exact settings;
  README owns membership/use; PORTAINER.md owns compact name/path metadata.
- Locally owned utilities use `tools/<tool>/`, with source, Dockerfile and manual
  build instructions together. No `images/` hierarchy or automated image builds.
  ThemeARR remains externally owned; do not copy it to invent a local tool.
- Cross-tool host workflows belong in `scripts/`; single-tool implementation stays
  with that tool. Temporary migrations use stable `canaries/<service>/` owners,
  explicit inactive/active status and MIGRATIONS history.
- `docs/` owns architecture, cross-cutting operations, evidence and decisions;
  link local owners instead of maintaining duplicate service/tool manuals. Keep
  unimplemented findings in [audits](docs/audits/service-improvements-2026-10.md);
  review and authorize them separately from current-state documentation.

## Version maintenance

Intentional upgrades require authorization and updates to the owning Compose/tool
source, stack/tool README version/build section, and [version inventory](docs/versions.md).
Compose/source is declaration authority; local tables are concise views; the
central page owns fleet evidence and discrepancies. Never change image tags just
to match documentation or Homelab Inventory. Record mutable/unpinned uncertainty.

## Preserve contracts and protected configuration

Inspect Git state and applicable instructions, preserve unrelated work, and keep
structural moves distinguishable from functional changes. Compare original and
moved service objects plus text/comments; avoid formatting or upgrades during moves.
Preserve `arrnet_managed`, IPs, names, bind modes, GPU/plugins and VPN identities.
Repository checkout prefixes and runtime state paths have different ownership.

Never read, search, source, parse, diff, hash, copy or expose protected root `.env`.
Exclude secrets, cookies, ignored SSO material and binaries from content searches.
Per-stack `.env.example` files contain only names from Compose with blank values;
real `.env` files remain ignored. Agents do not run Compose in this checkout.

Trace both Git consumers for path changes: repository sync/build/script/bind
paths, and the Git Compose metadata delivered to Luis. Do not inspect Portainer.
Publishing, manual builds and Portainer cutover are separate effects.

## Validate and deliver

Use static YAML parsing, exact service/allowed-delta comparisons, dependency/IP
checks, blank-template parity, documentation/version consistency and Bash syntax.
Run installed skill validators for applicable artifacts. Validate examples without
executing build commands, scripts, workflows, containers or runtime operations.

```bash
git diff --check -- . ':!.env' ':!**/.env'
```

When authorized to commit, stage explicit reviewed paths and use coherent commits
by purpose, ordered by dependencies. Review each staged diff and run relevant
checks before committing. Do not reset, clean, stash or squash unrelated work.
Push only with separate authorization: main publication triggers server sync.
Report commit order and remaining worktree changes; static checks do not establish
runtime availability or Luis's acceptance.
