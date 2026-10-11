# Acquisition

Acquire and process media while retaining the main VPN namespace and its consumers.

**Services:** `gluetun`, `qbittorrent`, `prowlarr`, `flaresolverr`, `ytdlp`, `themearr`, `makemkv`.

[Compose](compose.yml) owns exact configuration. [PORTAINER metadata](PORTAINER.md)
and the blank [environment contract](.env.example) are Luis's configuration handoff.
This is repository authority; deployment/cutover has not been performed here.

Gluetun owns the network namespace of qBittorrent, Prowlarr, FlareSolverr,
yt-dlp and ThemeARR. All depend on its health; never separate them from that
owner. MakeMKV processes the same `/data/fast/qbittorrent/downloads` tree exposed
by qBittorrent, giving it a useful home without a separate tool stack.

## Operational knowledge

- Gluetun retains `172.18.0.3`, tunnel device/NET_ADMIN, its own VPN inputs/state,
  and published ports 8081→8080, 9696 and 7681. Namespace consumers have no separate
  bridge interface; ThemeARR uses the 7681 web-terminal interface.
- Keep dependency `restart: true` where declared; yt-dlp intentionally has no such
  flag. A dependency restart flag differs from the container's restart policy and
  does not guarantee recovery from every VPN failure. Image healthchecks were not
  live-inspected; Compose defines no custom healthcheck.
- qBittorrent retains its remote VueTorrent LSIO mod, config/temp/watch and
  `/data/fast/qbittorrent` bind. The unused local bundle was removed. Luis explicitly
  removed deunhealth and its sole qBittorrent label; no replacement recovery was added.
- [yt-dlp](../../tools/yt-dlp/README.md) retains its idle shell and runtime music/
  downloads. Its cookie bind now expects the server-only file at
  `/home/devops/infra/media-stack/yt-dlp/cookies.txt`, not under `tools/`.
- [MakeMKV](../../tools/makemkv/README.md) retains no networking, restart `no` and
  its config/download binds. No new device passthrough or invocation model.
- [ThemeARR provenance](../../docs/tools/README.md#themearr-external-source) points
  to the external `animethemes-cli` recipe. Its current tag, API expression, output
  path and inherited image command are retained; no local source is imported.

## Service versions

| Service | Image | Version / tag |
| --- | --- | --- |
| `gluetun` | `qmcgaw/gluetun` | `v3.41.3` |
| `qbittorrent` | `lscr.io/linuxserver/qbittorrent` | `5.2.4_v2.0.15-ls479` |
| `prowlarr` | `lscr.io/linuxserver/prowlarr` | `2.6.5.5623-ls161` |
| `flaresolverr` | `ghcr.io/flaresolverr/flaresolverr` | `v3.5.2` |
| `ytdlp` | `ytdlp` | `latest` |
| `themearr` | `themearr` | `latest` |
| `makemkv` | `makemkv` | `latest` |

Declared tags from [Compose](compose.yml); [central versions](../../docs/versions.md)
owns provenance, fleet comparison and uncertainty.

## Shared contracts

**Prerequisite:** shared network `arrnet_managed` must exist before deployment.
Its lifecycle is externally managed / USER-owned by Luis; preserve declared IPs
and names. See [network provisioning](../../docs/operations/networking.md).
Luis owns Portainer activation and runtime validation.
See [architecture](../../docs/architecture.md),
[network/variable rules](../../docs/design/network-and-variables.md) and
[cutover handoff](../../docs/design/implementation-handoff.md).
