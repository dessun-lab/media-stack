# Subtitles

Acquire, place and align subtitles using the dedicated scraper VPN and cross-tool workflow.

**Services:** `gluetun_scrapers`, `flaresolverr_subs`, `nensaysubs`, `ffsubsync`.

[Compose](compose.yml) owns exact configuration. [PORTAINER metadata](PORTAINER.md)
and the blank [environment contract](.env.example) are Luis's configuration handoff.
Luis confirms this production stack is operational after cutover (operator update
recorded 2026-10-11); runtime operations remain his responsibility.

The dedicated scraper Gluetun owns the network namespace for FlareSolverr Subs
and NensaySubs. NensaySubs also depends on FlareSolverr Subs and uses its loopback
`http://127.0.0.1:8191/v1`. FFsubsync has no network dependency but participates
in the same [host workflow](../../docs/operations/subtitles.md) and media writes.

## Operational knowledge

- `gluetun_scrapers` retains `172.18.0.10`, separate SCRAPER VPN inputs, tunnel
  device/capability and runtime `arrstack/gluetun-scrapers`; no published ports.
- VPN/solver retain `unless-stopped`; NensaySubs and FFsubsync retain restart `no`.
  Dependency health/restart conditions remain unchanged.
- [NensaySubs](../../tools/nensaysubs/README.md) retains its idle image command,
  writable `/data/media` and `/runtime` bind. The host script requires the existing
  container named `nensaysubs`; no conversion to an on-demand job was made.
- [FFsubsync](../../tools/ffsubsync/README.md) retains `network_mode: none`, writable
  media and CLI entrypoint. The script uses a one-off `compose run --rm ffsubsync`
  with this unit's explicit Compose file, without starting its other services.
- Shared media writes may affect streaming even though lifecycle ownership is
  independent. Preserve host/container path agreement and operator configuration.

## Service versions

| Service | Image | Version / tag |
| --- | --- | --- |
| `gluetun_scrapers` | `qmcgaw/gluetun` | `v3.41.3` |
| `flaresolverr_subs` | `ghcr.io/flaresolverr/flaresolverr` | `v3.5.2` |
| `nensaysubs` | `nensaysubs` | `latest` |
| `ffsubsync` | `ffsubsync` | `latest` |

Declared tags from [Compose](compose.yml); [central versions](../../docs/versions.md)
owns provenance, fleet comparison and uncertainty.

## Shared contracts

**Shared infrastructure:** `arrnet_managed` is provisioned and in use per Luis.
It must remain available before any future deployment/recreation.
Its lifecycle is externally managed / USER-owned by Luis; preserve declared IPs
and names. See [network provisioning](../../docs/operations/networking.md).
Luis owns Portainer activation and runtime validation.
See [architecture](../../docs/architecture.md),
[network/variable rules](../../docs/design/network-and-variables.md) and
[completed migration record](../../docs/design/implementation-handoff.md).
