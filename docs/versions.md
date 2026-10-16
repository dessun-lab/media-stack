# Service and tool versions

[Documentation map](README.md)

**Existing fleet reconciliation: 2026-10-11; Bazarr addition: 2026-10-15.**
Compose/tool source is authoritative for what Media Stack declares. This page
owns version evidence and discrepancies; stack-local tables
are concise views of their Compose. It is not authorization to upgrade anything.

Fleet evidence: read-only `homelab-inventory/homelab-inventory.md`, commit
`32a87585bf6404ed4962d1ebf0945d97210ac26d`, `mty-srv01`, last inventoried
**2026-10-03 (America/Monterrey)**. Its tracked/deployed observations are distinct
from repository declarations and are not a fresh runtime inspection. No image
digest was queried for those existing services. Bazarr's new registry verification
is recorded separately below. `latest` is mutable, and `N/A` does not establish
an app version.

Luis confirms production is operational after cutover (update recorded 2026-10-11)
and Bazarr is deployed and functioning (update recorded 2026-10-15). These reports
establish operational status, not newly measured app versions or deployed image
digests. The dated fleet snapshot remains unchanged.

## Production declarations and fleet comparison

The owner links are the exact Compose sources for all **18 declared services**
(17 retained plus deployed Bazarr).
The application column comes from the dated fleet inventory, not image-tag inference.

| Service | Owner / declaration source | Image identity | Declared tag | Fleet application version | Comparison / uncertainty |
| --- | --- | --- | --- | --- | --- |
| `bazarr` | [subtitles](../stacks/subtitles/compose.yml) | `lscr.io/linuxserver/bazarr` | `v1.6.2-ls367` + digest below | Not inventoried | Luis reports it deployed and functioning; registry verified the declared image, but the deployed digest was not inspected. Newer build `ls368` exists; the declared pin remains `ls367`. |
| `jellyfin` | [streaming](../stacks/streaming/compose.yml) | `lscr.io/linuxserver/jellyfin` | `12.2ubu2604-ls53` | `12.1` | Luis-authorized declaration update; dated fleet still records `12.1ubu2604-ls50`. Deployment of the new tag is not asserted. |
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

Of 15 retained services present in fleet inventory, 14 image/tag strings match.
Jellyfin now declares `12.2ubu2604-ls53` per Luis's 2026-10-11 request; the dated
fleet records `12.1ubu2604-ls50` / application `12.1`. No new application version
or deployed digest was independently established.
FFsubsync and MakeMKV have no entries; their recipes are evidence below, not proof
of deployed versions. Fleet still lists `qmcgaw/deunhealth:latest` with application
`latest @ 37e5bd4`, while the applied restructuring removed it and Luis confirms
the old stack was removed. This is stale fleet evidence, not a pending cutover
action or proof of a remaining production container. No inventory was edited.

## Bazarr registry verification

On **2026-10-15**, public read-only LSCR metadata queries verified the
[tag listing](https://lscr.io/v2/linuxserver/bazarr/tags/list) and
[selected manifest](https://lscr.io/v2/linuxserver/bazarr/manifests/v1.6.2-ls367).
All ten tag-list pages were followed (9,441 tags); the stable `1.6.2` family
included these forms:

| Available form | Fixation / selection |
| --- | --- |
| `1.6.2`, `version-v1.6.2` | Application-version aliases; do not identify a LinuxServer rebuild. |
| `v1.6.2-ls366`, `v1.6.2-ls367`, `v1.6.2-ls368` | Explicit application + image build. `ls367` is the user-proposed candidate, not the newest listed build. |
| `amd64-…`, `arm64v8-…` counterparts | Architecture selection; does not make a tag immutable. |
| Build tag plus `@sha256:…` | Selected: readable app/build identity plus content-addressed fixation, stronger than a tag alone. |

The unprefixed `1.6.2-ls367` manifest returned HTTP 404; the leading `v` matters.
`version-v1.6.2` resolved to a different digest
(`sha256:fb0437897eaa7f96cf6ac42f795f8dcc208d34ec969eac5d366454e767304eb9`),
so it is not interchangeable with the selected build. Neither `latest`,
`development` nor a prerelease tag was selected. Both the
[upstream 1.6.2 release](https://github.com/morpheus65535/bazarr/releases/tag/v1.6.2)
and [LinuxServer ls367 release](https://github.com/linuxserver/docker-bazarr/releases/tag/v1.6.2-ls367)
report `prerelease: false` and `draft: false` in their public GitHub release metadata.

Exact Compose reference:

```text
lscr.io/linuxserver/bazarr:v1.6.2-ls367@sha256:c1673816321156b904af9b7e6415ea547ae9622e4acc375358f1e1eb9ed1e904
```

This is the OCI image-index digest, checked against both the registry response
header and SHA-256 of the manifest bytes. The index includes Linux `amd64` and
`arm64` images plus auxiliary descriptors; it retains platform selection without
floating the image contents. No image layers were pulled and Docker was not used.

A focused read-only check of `homelab-inventory/homelab-inventory.md` at
`194784db1303f54848566a04250387979f97fd43` (last inventoried 2026-10-15) found no
Bazarr entry. This dated fleet snapshot is incomplete for Bazarr and does not
refresh the older comparison for other services above. Registry evidence
establishes availability and the pin; Luis's report establishes deployment and
operation. See [Bazarr operations](operations/bazarr.md).

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
The inactive [Jellyfin canary](../canaries/jellyfin/compose.yml) retains historical
`lscr.io/linuxserver/jellyfin:12.1ubu2604-ls50`; it does not follow production
version updates automatically. See its migration log for history.

## Maintenance

For an authorized version change, update the owning Compose/tool definition, its
stack/tool README and this page together. Compare fleet evidence and record its
snapshot date; keep disagreement explicit instead of changing tags to make tables
agree. Manual rebuilds remain Luis-owned. No upgrade occurred during restructuring.
