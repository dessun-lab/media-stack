# Subtitles

Acquire, place and align subtitles using the dedicated scraper VPN and cross-tool workflow.

**Services:** `gluetun_scrapers`, `flaresolverr_subs`, `nensaysubs`, `ffsubsync`.

[Compose](compose.yml) owns exact configuration. [PORTAINER metadata](PORTAINER.md)
and the blank [environment contract](.env.example) are Luis's configuration handoff.
This is repository authority; deployment/cutover has not been performed here.

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

Use the existing external `arrnet_managed`; preserve declared IPs and names.
Luis owns network preparation, Portainer activation and runtime validation.
See [architecture](../../docs/architecture.md),
[network/variable rules](../../docs/design/network-and-variables.md) and
[cutover handoff](../../docs/design/implementation-handoff.md).
