# Streaming

Serve video and music independently of acquisition and library-management updates.

**Services:** `jellyfin`, `navidrome`.

[Compose](compose.yml) owns exact configuration. [PORTAINER metadata](PORTAINER.md)
and the blank [environment contract](.env.example) are Luis's configuration handoff.
This is repository authority; deployment/cutover has not been performed here.

Jellyfin and Navidrome have their own interfaces/state and no VPN or startup
dependency. Their common serving purpose justifies one update boundary; host,
media and plugin failures can still cross stack boundaries.

## Operational knowledge

- Jellyfin retains `172.18.0.2`, both 8096 host bindings, `/dev/dri`, video/render
  groups and read-only plugin DLLs from sibling `jellysun-ui/dist`. Config/cache/
  transcode/metadata stay under `/home/devops/runtime/jellysun`; media remains
  intentionally writable (history `20c4e96`). Literal image tag
  `12.1ubu2604-ls50` is preserved; do not normalize its version identity.
- Navidrome retains `172.18.0.8`, port 4533, `${PUID}:${PGID}` and runtime state
  under `/home/devops/runtime/arrstack/navidrome`. Its read-only music bind now
  expects `/home/devops/infra/media-stack/navidrome/music`; Luis prepares this
  server-only data before cutover. It differs from Lidarr's declared music source;
  no host alias or application mapping is inferred.
- Navidrome's external-auth trust expression, scanning/transcoding, metrics and
  sharing settings are unchanged. This repository does not audit active trust values.
- The [Jellyfin canary](../../canaries/jellyfin/README.md) is inactive and shares
  writable production state. It must be deliberately prepared before activation.

## Service versions

| Service | Image | Version / tag |
| --- | --- | --- |
| `jellyfin` | `lscr.io/linuxserver/jellyfin` | `12.1ubu2604-ls50` |
| `navidrome` | `deluan/navidrome` | `0.64.2` |

Declared tags from [Compose](compose.yml); [central versions](../../docs/versions.md)
owns provenance, fleet comparison and uncertainty.

## Shared contracts

Use the existing external `arrnet_managed`; preserve declared IPs and names.
Luis owns network preparation, Portainer activation and runtime validation.
See [architecture](../../docs/architecture.md),
[network/variable rules](../../docs/design/network-and-variables.md) and
[cutover handoff](../../docs/design/implementation-handoff.md).
