# Media Stack restructuring: approved implementation

[Documentation map](../README.md) · [Implementation handoff](implementation-handoff.md)

**Approved 2026-10-11; repository implementation and static validation complete.** Luis's implementation
brief supersedes the earlier assessment. Local commits are authorized by the completion addendum; Luis alone owns
publication and Portainer/runtime cutover.

## Selected ownership

Names are **streaming**, **management**, **acquisition**, **subtitles**. Streaming
clearly describes serving video/music; acquisition covers indexer access, downloads,
themes and downloaded-media processing. These follow Luis's strongest naming
options; four-unit membership is unchanged except explicitly removed deunhealth.
Exact membership belongs in each stack README, not duplicated across manuals.

| Owner | Classification rule |
| --- | --- |
| `stacks/<unit>/` | A useful independent production lifecycle. Each owns Compose, blank `.env.example`, README and compact Luis-owned PORTAINER metadata. |
| `tools/<tool>/` | Locally owned utility source, module/config, Dockerfile and manual build notes together. No `images/` hierarchy. |
| `scripts/` | Host workflows coordinating tools; the subtitle workflow stays here. |
| `canaries/<service>/` | Reusable temporary migration scenario, inactive by default in operating practice, with preparation and migration history. No automatic activation or version-folder cemetery. |
| `docs/` | Cross-cutting architecture, operations, evidence and decisions; ownership-local docs remain authoritative. |
| `.agents/skills/` | Existing local agent procedures; no planning framework. |
| `.github/workflows/` | Server-checkout synchronization, separate from Portainer; no image-build automation. |

## Implemented target

```text
README.md / CONTRIBUTING.md / AGENTS.md
docs/{README,architecture,evidence,versions}.md
docs/operations/{deployment,synchronization,subtitles}.md
docs/design/{restructuring,production-units,network-and-variables,implementation-handoff,canaries}.md
docs/tools/README.md
.agents/skills/{route-model-fit,optimize-agent-budget,execute-orchestrated-brief}/SKILL.md
.github/workflows/sync-repo.yml
stacks/
  streaming/{compose.yml,.env.example,README.md,PORTAINER.md}
  management/{compose.yml,.env.example,README.md,PORTAINER.md}
  acquisition/{compose.yml,.env.example,README.md,PORTAINER.md}
  subtitles/{compose.yml,.env.example,README.md,PORTAINER.md}
tools/
  ffsubsync/{Dockerfile,README.md}
  makemkv/{Dockerfile,README.md}
  nensaysubs/                 complete existing source/module/recipe/binary + README
  yt-dlp/{Dockerfile,README.md}
scripts/download-and-sync-subs.bash
canaries/
  README.md
  jellyfin/{compose.yml,README.md,MIGRATIONS.md}
```

Existing repository metadata remains. Protected root `.env` remains untouched;
only blank per-unit examples enter source control. ThemeARR remains externally
owned in `animethemes-cli`; do not import its recipe/source.

## Explicit changes and invariants

- Remove deunhealth and its exclusively owned qBittorrent label. This is an
  authorized functional removal, not a move. Remaining production coverage is 17.
- Delete local VueTorrent after the final reference check confirms the remote
  LSIO mod is the only active consumer mechanism; preserve that mod declaration.
- Move `jellyfin12` unchanged to `canaries/jellyfin/compose.yml`, preserving even
  its historical service/container identity. Record INACTIVE status and completed
  migration history. Shared writable state remains unsafe for concurrent use.
- Move four tool contexts unchanged, including the tracked NensaySubs executable.
  Manual builds remain Luis-triggered; no Compose build, CI, profiles or job conversion.
- Preserve service blocks, comments/style and semantics. Only the authorized
  deunhealth label removal and two infrastructure checkout bind prefixes differ.
- Replace `/home/devops/infra/arrstack/` with `/home/devops/infra/media-stack/`
  for this checkout's Navidrome music and yt-dlp cookie binds; keep subpaths/modes.
  Do not rename `/home/devops/runtime/arrstack` or sibling infrastructure paths.
- All four projects reference existing `arrnet_managed` externally, preserving
  eight IPs and both VPN namespaces; remove the unused duplicate default alias.
- Sync caller uses `github.event.repository.name`: the callee expects a short
  slug and uses it for GitHub and checkout destination. The new destination is
  intentional; the previous proposal to preserve the old checkout is superseded.
- Select subtitle Compose explicitly from script location; preserve all workflow
  modes, Docker exec/run behavior and arguments. No script/runtime execution here.

## Ordered checkpoints and acceptance

1. Reconcile this design and companions; record baseline Git/source state and
   perform the conditional VueTorrent no-consumer check before deleting anything.
2. Extract four Compose authorities by moving original text; remove deunhealth
   and its label, adapt only approved bind prefixes and shared network declaration.
3. Compare all 17 services structurally and textually, including comments; verify
   dependencies/IPs. **Only after this passes, remove root production Compose.**
4. Move complete tool contexts and inactive Jellyfin canary; preserve source bytes
   and binary. Add tool build notes and canary preparation/migration records.
5. Adapt script Compose selection and workflow slug; create blank per-stack
   contracts, local README and `media-stack-<unit>` PORTAINER metadata.
6. Reconcile current documentation, agent authority and contribution rules. Retire
   duplicated service manuals in favor of stack-local knowledge; preserve useful facts.
7. Maintain central version provenance and stack-local declared-version tables;
   compare the read-only fleet inventory without changing image tags.
8. Validate service parity against baseline, allowed deltas, text/comments, paths,
   blank variables, YAML/workflow contract, manual commands, Bash, skills, links,
   ignore rules and final diff/worktree. No protected `.env` or runtime effects.
9. Report repository completion and exact Luis-owned operational handoff. Do not
   push, contact Portainer, SSH to server, build or create runtime resources.
10. Under the completion addendum, review/stage explicit paths in coherent local
    commits by purpose; validate each group, report hashes/order and remaining dirt.
    No catch-all commit or push is authorized.
