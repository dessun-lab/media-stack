# Management

Coordinate media libraries and requests independently of serving and VPN maintenance.

**Services:** `radarr`, `sonarr`, `lidarr`, `jellyseerr`.

[Compose](compose.yml) owns exact configuration. [PORTAINER metadata](PORTAINER.md)
and the blank [environment contract](.env.example) are Luis's configuration handoff.
This is repository authority; deployment/cutover has not been performed here.

These four services own requests and movie/series/music library automation.
They have independent bridge interfaces and no Compose startup dependencies.
Prowlarr stays in acquisition because its VPN namespace is a hard dependency.

## Operational knowledge

| Service | Interface | Persisted ownership |
| --- | --- | --- |
| `radarr` | `172.18.0.4`, 7878 | Runtime `arrstack/radarr`: config, movies, downloads. |
| `sonarr` | `172.18.0.5`, 8989 | Runtime `arrstack/sonarr`: config, tv, shows, downloads. |
| `lidarr` | `172.18.0.9`, 8686 | Runtime `arrstack/lidarr`: config, music, downloads. |
| `jellyseerr` | `172.18.0.7`, 5055 | Runtime `arrstack/jellyseerr` → `/app/config`; image is Seerr, container name stays `jellyseerr`. |

Runtime prefix is `/home/devops/runtime`; all services retain `unless-stopped`.
Exact image tags and mount modes belong to Compose. Inventory dated 2026-10-03
records Lidarr `3.1.0.4875` (`ls26`) behind the mutable `latest` tag, not a source pin.
Managers use Prowlarr/qBittorrent through application configuration. Preserve
endpoints/credentials and request-to-streaming integration at Luis's cutover;
these relationships do not require a common Compose project. Separate mount
sources do not prove shared downloads, aliases or hardlink behavior.

## Service versions

| Service | Image | Version / tag |
| --- | --- | --- |
| `radarr` | `lscr.io/linuxserver/radarr` | `6.4.4.10685-ls317` |
| `sonarr` | `lscr.io/linuxserver/sonarr` | `4.0.20.3014-ls325` |
| `lidarr` | `lscr.io/linuxserver/lidarr` | `latest` |
| `jellyseerr` | `ghcr.io/seerr-team/seerr` | `v3.5.0` |

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
