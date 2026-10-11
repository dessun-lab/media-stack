# Implementation evidence and validation

[Documentation map](README.md) · [Approved design](design/restructuring.md)

## Historical implementation scope and authority

This page retains the checks/effect limits from the original restructuring. Luis
has since confirmed cutover complete and all four stacks operational (update
recorded 2026-10-11); see the [completed migration record](design/implementation-handoff.md).
Subsequent service-header edits are comments only, so the original byte/comment
parity results below describe the original implementation checkpoint.

Repository restructuring implemented **2026-10-11** from Luis's approved brief,
including the version-documentation and local-commit addenda. Runtime effects,
Portainer access, server SSH, builds, secret handling and pushing remain excluded.
The existing documentation/skills worktree was reconciled as intended prior work;
no unrelated changes were reset, stashed or discarded.

Baseline Git HEAD: `5956f4cbaccebe8d03b1d23eead73aaf98d984f4` on `main`.
Before edits, named non-secret tracked assets and authorized Markdown were copied
to a task-local temporary baseline for comparison. Root `.env`, cookies and the
ignored SSO file were never read, copied, hashed, parsed, diffed or used as input.
The SSO file was discovered through filename metadata, then renamed unchanged
with Luis's explicit permission and confirmed ignored at its new location.

## Evidence sources

| Source | Supported conclusion |
| --- | --- |
| Baseline root Compose and current four stack files | 18 old services minus authorized deunhealth removal equals 17 retained services exactly once. Existing images/tags, namespace ownership and other service semantics are preserved except enumerated differences below. |
| Final narrow VueTorrent reference check | qBittorrent uses `ghcr.io/vuetorrent/vuetorrent-lsio-mod:latest`; no active repository mount/caller consumes the local bundle. All bundle files were tracked; conditional deletion proceeded. History `f456a76`, `700cf79`, `3c0c228` supports the earlier switch from local bind to remote mod. |
| deunhealth references | Sole integration label was `deunhealth.restart.on.unhealthy=true` on qBittorrent. Removed with the service; TZ remains required elsewhere, and literal service-only settings/socket disappeared with its block. No unrelated qBittorrent behavior changed. |
| Jellyfin source/history | Canary definition moved byte-for-byte. [MIGRATIONS](../canaries/jellyfin/MIGRATIONS.md) records dated creation/state/promotion commits separately from Luis's successful validation/inactive-status evidence. |
| Four local tool contexts | Eight tracked assets moved byte-for-byte with modes preserved, including NensaySubs native executable. No Dockerfile or application-source modernization. |
| `dl-actions` at `8e8cb5b50bbb84c83b832a03c98129615d733e32` | Narrow callee read confirms short slug controls GitHub URL and checkout directory; dynamic repository name fits both supported triggers. No reference-repository edit or workflow execution. |
| `homelab-inventory` at `32a87585bf6404ed4962d1ebf0945d97210ac26d` | Dated fleet tags/app versions compared in [versions](versions.md). Fifteen retained tags agree; FFsubsync/MakeMKV absent; removed deunhealth still inventoried. No fleet edits. |
| External `animethemes-cli` at `aa949a4b7b4850a69ea29b054a0bd7e342809191` | Prior narrow recipe/interface evidence establishes ThemeARR's matching external owner; [tool index](tools/README.md) preserves its provenance limits. Source not imported; deployed binary revision remains unverified. |

Repository source is declaration authority; Luis supplies operating facts; fleet
inventory supplies dated observations. No live parity, runtime success or remote
mutable-ref freshness is inferred from local evidence.

## Exhaustive intentional differences

| Difference | Scope and preservation check |
| --- | --- |
| Functional removal | deunhealth service and exclusive qBittorrent label only. No substitute recovery mechanism. |
| Shared network ownership | Each production project uses external `arrnet_managed`; unused default alias and project-owned IPAM omitted. Eight existing service addresses and all namespace consumers unchanged. No actual network creation/change. |
| Infrastructure checkout prefix | Only Navidrome music and yt-dlp cookies change `/home/devops/infra/arrstack/` to `/home/devops/infra/media-stack/`; subpaths/modes preserved. Runtime and sibling infrastructure paths unchanged. |
| Compose source ownership | Four files replace root authority; nonblank service/comment lines preserved, except the removed block/label and two bind prefixes. Whitespace-only separator lines normalized; no comment/style rewrite. |
| Script discovery | Added script-directory-derived subtitle Compose path and `-f` argument only. All modes, exec/run behavior and arguments preserved; script not executed. |
| Sync identity | Only the reusable input changes to `github.event.repository.name` plus final newline; triggers/callee/server/secrets behavior unchanged. Intended new server destination is a Luis-owned preparation handoff. |
| Filesystem organization | Complete tools and unchanged canary moved; unused VueTorrent removed; old locations retired. Ignored SSO moved separately by explicit permission, without content access. |
| Documentation/config hygiene | Local README/PORTAINER metadata, blank examples, version inventory, navigable cross-cutting docs and ignore-rule protection. No operational values or automatic builds introduced. |

## Validation results

A deterministic temporary audit used plain YAML parsing, never Docker/Compose
interpolation or protected configuration. It checked:

- Exact old/new object equality for all 17 retained services after only the
  deunhealth-label and two bind-prefix adjustments; images/tags unchanged.
- No missing/extra/duplicate production services; all eight `depends_on` edges
  and every VPN namespace reference remain inside the owning unit.
- All eight static IPs map to the same service; four external-network declarations
  match the shared contract. Canary Compose is byte-identical to its baseline.
- Every nonblank service/comment line is accounted for after enumerated removals;
  tool recipes/source/binary bytes and file modes match baseline.
- Four blank examples exactly match Compose inputs: 26 distinct names, 33 entries
  across units. Real `.env` and `.env.*` paths remain ignored; examples are allowed.
  Protected root `.env` was checked only for ignore status, never content.
- Every production service appears in the central version table and its stack's
  table with exact declared image/tag. Fleet comparison records missing/removed
  entries without silently upgrading or modifying source to match inventory.
- Workflow YAML and caller/callee string contract pass static checks; both push
  and dispatch repository-name fixtures resolve the intended new destination.
  `actionlint` is not installed; no remote workflow execution is claimed.
- Bash syntax and all documented manual build blocks pass without execution.
  Commands match new contexts/tags, including MakeMKV's recipe argument and
  yt-dlp's empty-context build. No source build was necessary for unchanged tools.
- Local Markdown links/anchors, fenced blocks, final newlines and whitespace are
  passed for 34 Markdown artifacts, 187 local links/anchors and seven Bash examples.
- All three skill frontmatter checks and deterministic local quality checks passed
  (88.5 / 88.0 / 88.5). Optional metadata/section advisories remain non-blocking;
  versions stay in schema-compatible metadata. Scoped `git diff --check` passed.
  Each local commit uses reviewed explicit index paths and a staged-diff review;
  no push is authorized.

## Agent-procedure provenance

The three local skills are ADAPT migrations from read-only Experiment Platform
`28e32d3839280c95091571b3f178a1b319d36be3`: route-model-fit (source 1.5.0),
optimize-agent-budget (1.1.0), execute-orchestrated-brief (1.1.0). The latter is the
verified source spelling; no alias was installed. Retained task-fit reasoning,
bounded investigation, concise receipts and effect/handoff boundaries. Removed
price/advisor infrastructure, scientific/product assumptions and Spec Kit rules.
No new Experiment Platform inspection or modification was needed in implementation.

## Acceptance boundary

Original static acceptance did not establish runtime deployment. Luis's subsequent
operator update establishes the completed cutover: old stack removed, four stacks
operational, network provisioned and Portainer variable contracts populated.
No agent runtime inspection occurred. Luis retains future operational authority;
the canary remains INACTIVE. See the [completed migration record](design/implementation-handoff.md).
