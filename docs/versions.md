# Service and tool versions

[Documentation map](README.md)

**Reconciled 2026-10-11.** Compose/tool source is authoritative for what Media Stack
declares. This page owns version evidence and discrepancies; stack-local tables
are concise views of their Compose. It is not authorization to upgrade anything.

Fleet evidence: read-only `homelab-inventory/homelab-inventory.md`, commit
`32a87585bf6404ed4962d1ebf0945d97210ac26d`, `mty-srv01`, last inventoried
**2026-10-03 (America/Monterrey)**. Its tracked/deployed observations are distinct
from repository declarations and are not a fresh runtime inspection. No image
digest was queried. `latest` is mutable, and `N/A` does not establish an app version.

Luis confirms production is operational after cutover (update recorded 2026-10-11).
That report establishes deployment status, not newly measured app versions or
custom-image digests. The dated fleet snapshot remains unchanged.

## Production declarations and fleet comparison

The owner links are the exact Compose sources for all **17 retained services**.
The application column comes from the dated fleet inventory, not image-tag inference.

| Service | Owner / declaration source | Image identity | Declared tag | Fleet application version | Comparison / uncertainty |
| --- | --- | --- | --- | --- | --- |
| `jellyfin` | [streaming](../stacks/streaming/compose.yml) | `lscr.io/linuxserver/jellyfin` | `12.1ubu2604-ls50` | `12.1` | Literal tag and inventory label retained; no inferred upstream renumbering. |
| `navidrome` | [streaming](../stacks/streaming/compose.yml) | `deluan/navidrome` | `0.64.2` | `0.64.2` | Same declared image/tag; dated observation. |
| `radarr` | [management](../stacks/management/compose.yml) | `lscr.io/linuxserver/radarr` | `6.4.4.10685-ls317` | `6.4.4.10685` | Same declared image/tag; dated observation. |
| `sonarr` | [management](../stacks/management/compose.yml) | `lscr.io/linuxserver/sonarr` | `4.0.20.3014-ls325` | `4.0.20.3014` | Same declared image/tag; dated observation. |
| `lidarr` | [management](../stacks/management/compose.yml) | `lscr.io/linuxserver/lidarr` | `latest` | `3.1.0.4875` (`ls26`) | Same mutable `latest`; observed app/build does not pin this declaration. |
| `jellyseerr` | [management](../stacks/management/compose.yml) | `ghcr.io/seerr-team/seerr` | `v3.5.0` | `3.5.0` | Same declared image/tag; dated observation. |
| `gluetun` | [acquisition](../stacks/acquisition/compose.yml) | `qmcgaw/gluetun` | `v3.41.3` | `3.41.3` | Same declared image/tag; dated observation. |
| `qbittorrent` | [acquisition](../stacks/acquisition/compose.yml) | `lscr.io/linuxserver/qbittorrent` | `5.2.4_v2.0.15-ls479` | `5.2.4` | Same declared image/tag; dated observation. |
| `prowlarr` | [acquisition](../stacks/acquisition/compose.yml) | `lscr.io/linuxserver/prowlarr` | `2.6.5.5623-ls161` | `2.6.5.5623` | Same declared image/tag; dated observation. |
| `flaresolverr` | [acquisition](../stacks/acquisition/compose.yml) | `ghcr.io/flaresolverr/flaresolverr` | `v3.5.2` | `3.5.2` | Same declared image/tag; dated observation. |
| `ytdlp` | [acquisition](../stacks/acquisition/compose.yml) | `ytdlp` | `latest` | N/A | Same custom mutable tag; deployed source/app version unrecorded. |
| `themearr` | [acquisition](../stacks/acquisition/compose.yml) | `themearr` | `latest` | N/A | Same custom mutable tag; deployed source/app version unrecorded. |
| `makemkv` | [acquisition](../stacks/acquisition/compose.yml) | `makemkv` | `latest` | Not inventoried | No dated fleet row; exact deployed tool version remains unrecorded. |
| `gluetun_scrapers` | [subtitles](../stacks/subtitles/compose.yml) | `qmcgaw/gluetun` | `v3.41.3` | `3.41.3` | Same declared image/tag; dated observation. |
| `flaresolverr_subs` | [subtitles](../stacks/subtitles/compose.yml) | `ghcr.io/flaresolverr/flaresolverr` | `v3.5.2` | `3.5.2` | Same declared image/tag; dated observation. |
| `nensaysubs` | [subtitles](../stacks/subtitles/compose.yml) | `nensaysubs` | `latest` | N/A | Same custom mutable tag; deployed source/app version unrecorded. |
| `ffsubsync` | [subtitles](../stacks/subtitles/compose.yml) | `ffsubsync` | `latest` | Not inventoried | No dated fleet row; exact deployed tool version remains unrecorded. |

All 15 retained services present in fleet inventory have matching image/tag strings.
FFsubsync and MakeMKV have no entries; their recipes are evidence below, not proof
of deployed versions. Fleet still lists `qmcgaw/deunhealth:latest` with application
`latest @ 37e5bd4`, while the applied restructuring removed it and Luis confirms
the old stack was removed. This is stale fleet evidence, not a pending cutover
action or proof of a remaining production container. No inventory was edited.

## Custom tool recipe versions

| Tool / source | Image tag | Base / tool version evidence | Limit |
| --- | --- | --- | --- |
| [FFsubsync](../tools/ffsubsync/README.md), local Dockerfile | `ffsubsync:latest` | `python:3.12-slim`; pip ffsubsync unpinned, distro FFmpeg | App version not pinned or fleet-recorded. |
| [MakeMKV](../tools/makemkv/README.md), local Dockerfile ARG | `makemkv:latest` | `ubuntu:24.04`; `MAKEMKV_VERSION=1.18.3` default/manual recipe | Build argument can override; actual deployed version not established. |
| [NensaySubs](../tools/nensaysubs/README.md), Dockerfile/go.mod | `nensaysubs:latest` | `golang:1.26-alpine`; module Go `1.26.3`, CGO/Tesseract/Leptonica | Go version is toolchain, not app release. No app release pin; tracked binary retained without claiming deployed parity. |
| [yt-dlp](../tools/yt-dlp/README.md), local Dockerfile | `ytdlp:latest` | `python:3.12-alpine`; pip yt-dlp unpinned | Fleet records `N/A`; no reproducible app pin claimed. |
| [ThemeARR](tools/README.md#themearr-external-source), external `animethemes-cli` | `themearr:latest` | Matching external recipe at `aa949a4`: `debian:bookworm-slim`, Go source `1.26.2`; copied dist metadata `0.0.0-SNAPSHOT-cc506a1` (2026-05-09) | Dockerfile copies a prebuilt binary; metadata/source match does not prove the deployed image revision. |

qBittorrent also declares `ghcr.io/vuetorrent/vuetorrent-lsio-mod:latest` as its
Docker mod. This remote mutable dependency is preserved; the removed local bundle
was version `2.33.0` and is not a statement of the remote mod's current version.
The inactive [Jellyfin canary](../canaries/jellyfin/compose.yml) retains the same
literal Jellyfin image/tag as production; see its migration log for history.

## Maintenance

For an authorized version change, update the owning Compose/tool definition, its
stack/tool README and this page together. Compare fleet evidence and record its
snapshot date; keep disagreement explicit instead of changing tags to make tables
agree. Manual rebuilds remain Luis-owned. No upgrade occurred during restructuring.
