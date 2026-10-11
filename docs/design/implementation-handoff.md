# Completed Media Stack migration

[Design overview](restructuring.md) · [Current deployment](../operations/deployment.md)

**APPLIED — operator confirmation recorded 2026-10-11.** Luis reports the previous
monolithic Portainer stack removed, the four production stacks created, shared
external network provisioned, environment contracts populated and production
services up again. The new architecture is operational. This record replaces the
completed implementation handoff; it is not a remaining migration checklist.

Repository identity is `dessun-lab/media-stack`; origin uses
`git@github-lab:dessun-lab/media-stack.git`. Agent static validation and Luis's
runtime acceptance are separate evidence. No agent inspected Portainer or runtime.
The report does not assert exact Portainer display names, image digests, sync-job
history or the disposition of an old server checkout.

## Exact ownership and path changes

| Before | Repository result / operational contract |
| --- | --- |
| Root `docker-compose.yml` | Four `stacks/{streaming,management,acquisition,subtitles}/compose.yml` authorities; no aggregate wrapper. See each adjacent PORTAINER.md. |
| Four root tool directories | Complete assets under `tools/<tool>/`; manual build context documented locally; no source/recipe modernization. |
| `jellyfin12/docker-compose.yml` | `canaries/jellyfin/compose.yml`, unchanged semantics and historical container name; inactive, no concurrent-state safety claim. |
| `vuetorrent/` | Removed after no active repository consumer confirmed; remote LSIO mod retained. |
| deunhealth service + qBittorrent label | Explicit functional removal. No replacement recovery mechanism or unrelated qBittorrent edit. |
| Host subtitle script | Same `scripts/` location; explicit subtitle Compose path derived from `BASH_SOURCE`, keeping `exec` and `run --rm` models. |
| `/home/devops/infra/arrstack/navidrome/music` | `/home/devops/infra/media-stack/navidrome/music`, still read-only `/music`. |
| `/home/devops/infra/arrstack/yt-dlp/cookies.txt` | `/home/devops/infra/media-stack/yt-dlp/cookies.txt`, same subpath and mount mode. This server-only credential does NOT follow the Dockerfile into `tools/`. |
| Sync caller `repository: arrstack` | `repository: ${{ github.event.repository.name }}`. Existing callee constructs both GitHub slug and `${DEPLOY_PATH}/${REPOSITORY}`; intended destination now matches the new slug. |

Only those two infrastructure binds reference the renamed checkout. All
`/home/devops/runtime/arrstack`, `/home/devops/runtime/jellysun`, `/data` and sibling
`/home/devops/infra/jellysun-ui` paths remain unchanged. The previous design's
old-checkout compatibility/decoupled-input proposal is superseded by Luis's explicit
new-destination decision; no `dl-actions` change is necessary.

## Durable recovery boundaries

Current stack paths and blank variable contracts live with their stack owners;
[deployment](../operations/deployment.md), [networking](../operations/networking.md)
and [synchronization](../operations/synchronization.md) own ongoing operations.
Luis controls future publication, runtime maintenance and rollback. Clone/pull
cannot restore server-only music, cookies or mutable application data. Preserve
those separately; no secret values belong in this record.

The old Compose owned its network: restoring old Git files alone is not an
operational rollback and may reintroduce network ownership/deunhealth. Never run
old and new owners of the same names, addresses or writable state concurrently.
The Jellyfin canary remains inactive; its earlier migration is separate from the
completed production-stack cutover and future activation requires isolation.
