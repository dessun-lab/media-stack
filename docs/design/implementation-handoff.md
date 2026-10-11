# Implementation and Luis-owned operational handoff

[Design overview](restructuring.md) · [Deployment](../operations/deployment.md)

Approved 2026-10-11. Repository identity is `dessun-lab/media-stack`; origin remains
`git@github-lab:dessun-lab/media-stack.git`. Local checkout already has the new name.
No local directory rename, remote publication or runtime action is part of this task.

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

## Publication and server preparation — Luis

A later authorized main push invokes sync. With the operator's infrastructure
`DEPLOY_PATH`, it targets `/home/devops/infra/media-stack`; it may clone there
instead of updating the old checkout. A clone does not migrate server-only music,
cookies, configuration, aliases or schedules. Luis prepares these exact two bind
sources and script/config invocation context before switching consumers. No secret
value is copied or reconstructed here. The old checkout is not remotely removed.

## Portainer and runtime — Luis only

Use the repository `dessun-lab/media-stack` and the four stack-local PORTAINER.md
files for recommended `media-stack-<unit>` names, exact Compose paths and blank
environment contracts. Luis supplies operational values, configures Git/updates,
creates/replaces stacks, deploys and controls rollback. Agents do not access or
inspect Portainer; its settings are not repository implementation gates.

Luis preserves/prepares `arrnet_managed` and existing addresses before activation.
Prevent old-project cleanup from deleting the shared network; transfer fixed
container names/IPs/state ownership once. Prepare backups and interruption/rollback
procedures; never activate old and new writers simultaneously. Ensure existing
custom tags are available; builds remain manual. Retire any old deunhealth deployment
as part of Luis's cutover, not by assuming repository deletion stopped it.

Runtime acceptance covers streaming/plugins/GPU, management/application endpoints,
both VPN namespaces/health dependencies, acquisition and subtitle paths. Reverting
Git alone cannot restore deployment ownership or changed data. The Jellyfin canary
stays inactive; isolation/activation is a separately owned future migration.
