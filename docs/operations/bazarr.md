# Bazarr operations

[Documentation map](../README.md) · [Subtitle stack](../../stacks/subtitles/README.md) ·
[Version evidence](../versions.md#bazarr-registry-verification)

Luis confirms Bazarr is deployed and functioning in Media Stack (operator update
recorded 2026-10-15). It manages downloaded Spanish and English subtitles for
Radarr/Sonarr libraries so Jellyfin can use sidecar tracks. This status is operator
evidence; repository agents have not inspected Portainer, Bazarr's UI or the host.
The repository does not implement subtitle generation or translation.

## Network and service ownership

[Compose](../../stacks/subtitles/compose.yml) owns the exact image, binds and
network settings. Bazarr runs in the subtitles stack with
`network_mode: service:gluetun_scrapers`; it shares Gluetun's `172.18.0.10`
interface, VPN tunnel and firewall. It starts after Gluetun is healthy and follows
explicit Compose restarts of its network owner. Provider traffic uses the scraper
VPN. A VPN outage may interrupt downloads and searches.

Gluetun publishes Bazarr's UI on `${BAZARR_HOST}:6767` and allows the port through
its default-interface firewall with `FIREWALL_INPUT_PORTS=6767`. The UI is reachable
on the host interface Luis configured; this is local ingress, not VPN-provider
port forwarding. Keep Bazarr authentication and the host-interface binding under
Luis's operational control.

`FIREWALL_OUTBOUND_SUBNETS=172.18.0.4/32,172.18.0.5/32` permits local access from
the scraper namespace to Radarr and Sonarr. These IP-wide exceptions apply to
**every** consumer of that namespace. They do not exempt public destinations from
the VPN route. If either manager's static address changes, update the Gluetun
exception and Bazarr's application connection together. Gluetun's
[namespace](https://github.com/qdm12/gluetun-wiki/blob/main/setup/connect-a-container-to-gluetun.md)
and [firewall](https://github.com/qdm12/gluetun-wiki/blob/main/setup/options/firewall.md)
documentation explains these settings.

Recreating or changing Gluetun can affect Bazarr, NensaySubs and FlareSolverr Subs.
Coordinate the shared namespace's lifecycle in Portainer. The Compose dependency
restart setting does not replace recovery checks after an unexpected failure.

## Media and persistent configuration

| Host bind | Bazarr path | Purpose |
| --- | --- | --- |
| `/home/devops/runtime/arrstack/bazarr` | `/config` | Persistent Bazarr settings, database and credentials. |
| `/home/devops/runtime/arrstack/radarr/movies` | `/movies` | Radarr movie files and writable subtitle sidecars. |
| `/home/devops/runtime/arrstack/sonarr/tv` | `/tv` | Sonarr TV files and writable sidecars. |
| `/home/devops/runtime/arrstack/sonarr/shows` | `/shows` | Additional Sonarr series files and writable sidecars. |

`BAZARR_HOST`, `PUID`, `PGID` and `TZ` are Portainer inputs documented as blank
names in the [subtitle stack template](../../stacks/subtitles/.env.example).
Bazarr uses the managers' UID/GID and `UMASK=002` convention. Its `/config` and
library mounts must remain writable to that identity; do not solve a permission
failure with a broad recursive ownership change. API keys, provider credentials
and language-profile settings belong in Bazarr's persistent `/config` and UI, not
in Git. The repository records their intended role, not their active values.

The declared manager endpoints are:

| Integration | Bazarr host / IP | Port | Address |
| --- | --- | --- | --- |
| Radarr | `172.18.0.4` | `7878` | `http://172.18.0.4:7878` |
| Sonarr | `172.18.0.5` | `8989` | `http://172.18.0.5:8989` |

Each integration uses its application's API key in Bazarr. The container paths
match the managers' paths, so path mappings are unnecessary for `/movies`, `/tv`
and `/shows`. Bazarr's UI owns the Spanish/English language profile, provider
accounts and subtitle placement setting; preserve both languages when changing
profiles and keep sidecars beside media for Jellyfin. The
[official Bazarr setup guide](https://wiki.bazarr.media/Getting-Started/Setup-Guide/)
describes those controls.

## Operation and troubleshooting

For a missing subtitle, check the item's Bazarr profile, monitored status,
provider availability and History; then check manager connectivity and media
permissions. A provider without a suitable result is different from a failed API
connection. Jellyfin recognizes external language suffixes such as
`Example.es.srt` and `Example.en.srt` beside `Example.mkv`; see its
[external-track rules](https://jellyfin.org/docs/general/server/media/movies/#external-subtitles-and-audio-tracks).
Refresh the affected library if a new sidecar does not appear in Jellyfin.

Jellyfin mounts `/data/media`, while the managers and Bazarr use the host paths
above. The repository alone does not establish whether these are aliases for the
same underlying media. Luis's report establishes that Bazarr is functioning;
confirm the same-file relationship when changing mounts or investigating a
Jellyfin visibility issue. Do not introduce path mappings or host aliases without
establishing that relationship.

NensaySubs and FFsubsync remain separate manually operated tools. Avoid having
multiple tools overwrite one sidecar concurrently. No automated alignment,
AI subtitle generation or translation is part of Bazarr's repository definition.
