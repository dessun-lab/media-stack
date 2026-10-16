# Bazarr: Radarr/Sonarr subtitles

[Documentation map](../README.md) · [Subtitle stack](../../stacks/subtitles/README.md) ·
[Version evidence](../versions.md#bazarr-registry-verification)

**Repository addition: 2026-10-15; deployment and application acceptance pending
Luis.** Bazarr manages downloaded Spanish and English subtitles for the libraries
owned by Radarr/Sonarr. No AI generation, translation, additional subtitle service,
or change to the existing NensaySubs/FFsubsync workflow is included.

## Service contract

[Compose](../../stacks/subtitles/compose.yml) owns the exact image and configuration.
Per Luis's egress policy, Bazarr uses `network_mode: service:gluetun_scrapers`
and shares its existing `172.18.0.10` interface, VPN tunnel and firewall. It waits
for Gluetun health at startup and declares `depends_on.restart: true`, matching
the existing namespace consumers. Internet provider traffic is intended to use
the scraper VPN; there is no independent Bazarr bridge interface or direct-egress
fallback. Radarr/Sonarr remain application dependencies across stacks, with no
cross-project `depends_on`.

Gluetun publishes `${BAZARR_HOST}:6767:6767` and permits `FIREWALL_INPUT_PORTS=6767`
on its default interface. This is local UI ingress, not VPN-provider port forwarding.
`FIREWALL_OUTBOUND_SUBNETS=172.18.0.4/32,172.18.0.5/32` allows local manager access
without exempting the entire Docker subnet or public internet from the VPN route.
Those two IP exceptions apply to **every consumer** of the scraper namespace,
not just Bazarr, and are IP-wide rather than API-port-specific. Existing manager
addresses are therefore part of this integration contract.

These settings follow Gluetun's [namespace sharing](https://github.com/qdm12/gluetun-wiki/blob/main/setup/connect-a-container-to-gluetun.md)
and [firewall options](https://github.com/qdm12/gluetun-wiki/blob/main/setup/options/firewall.md).
Use manager IPs below to avoid relying on Docker-name resolution through the VPN
resolver. No public DNS/private-address exception is added.

Updating/recreating the VPN owner can affect Bazarr and the existing subtitle
tools. Luis must recreate namespace consumers with their owner when necessary;
the dependency restart setting is not a general watchdog for runtime failures.
No management/streaming redeploy is required by this repository change. VPN
routing does not make logged-in provider accounts anonymous or hide the locally
published UI; authenticate it and keep the intended host-interface binding.

The [LinuxServer image](https://docs.linuxserver.io/images/docker-bazarr/) supports
UID/GID ownership, persistent `/config`, library mounts and web setup on port
`6767`. This service follows the managers' `PUID`, `PGID`, `TZ`, `UMASK=002` and
`unless-stopped` conventions. The exact stable build and multi-platform digest
are recorded in the central version inventory.

| Host bind | Bazarr path | Ownership / access |
| --- | --- | --- |
| `/home/devops/runtime/arrstack/bazarr` | `/config` | New Bazarr-only persistent configuration, database and credentials; writable. |
| `/home/devops/runtime/arrstack/radarr/movies` | `/movies` | Same movie bind/path as Radarr; writable for sidecar subtitles. |
| `/home/devops/runtime/arrstack/sonarr/tv` | `/tv` | Same TV bind/path as Sonarr; writable. |
| `/home/devops/runtime/arrstack/sonarr/shows` | `/shows` | Same additional series bind/path as Sonarr; writable. |

No download directory, other application's configuration, Docker socket or GPU
is mounted. Existing media ownership is preserved; Luis verifies the selected
UID/GID can traverse/read media and create sidecars. A umask does not grant access
to an otherwise unwritable directory. Do not recursively change library ownership
as an installation shortcut.

## Luis-owned provisioning and setup

1. Before deploying, verify `arrnet_managed` exists and port `6767` is free on
   the selected host interface and within the scraper namespace. No new IP is
   needed. Verify the manager-IP exceptions do not overlap the configured VPN
   tunnel range. Prepare the new `/config` host directory with the media-owner
   UID/GID. Runtime configuration was not inspected by agents.
2. Populate new subtitle-stack inputs `BAZARR_HOST`, `PUID`, `PGID`; retain `TZ`
   and all existing scraper inputs. `BAZARR_HOST` is a **host bind IP address**,
   not a URL or port. Choose the intended private/reachable interface; do not leave
   it blank and accidentally publish on all interfaces. Use the same UID/GID as
   the managers. The blank [template](../../stacks/subtitles/.env.example) contains
   names only. Root `.env` remains protected.
3. Luis deploys the updated `stacks/subtitles/compose.yml` through Portainer when
   ready, including Gluetun's new port/firewall configuration and its consumers.
   Open `http://<BAZARR_HOST>:6767` and configure authentication before
   allowing other users access. Persistent credentials stay in `/config`, outside
   this repository; do not paste API keys into Git, logs or documentation.
4. Enable both integrations in Bazarr, supply each application's own API key
   through the UI, and successfully test/save each connection:

   | Application | Host / IP | Port | Local endpoint outside the VPN tunnel |
   | --- | --- | --- | --- |
   | Radarr | `172.18.0.4` | `7878` | `http://172.18.0.4:7878` |
   | Sonarr | `172.18.0.5` | `8989` | `http://172.18.0.5:8989` |

   These are the managers' existing Compose IPs; `localhost` is the shared scraper
   namespace, not either manager.
   Direct endpoints assume no application URL base/TLS override. If Luis configured
   one, match that application setting. The declared library paths match exactly,
   so leave path mappings empty for `/movies`, `/tv` and `/shows`.
5. Enable Spanish and English, then create an `ES + EN` profile containing both.
   Keep Single Language off. Do not set a cutoff that stops after either language,
   or exclude matching audio, when both subtitles are wanted. Apply the profile
   to new movies/series by default and explicitly assign it to existing items.
6. Enable download providers supporting the requested languages and enter any
   account credentials in Bazarr. Provider availability, quotas and credentials
   are operator inputs. Keep Whisper/AI and translation integrations disabled.
   Save subtitles `AlongSide Media File`; leave custom post-processing disabled
   for this initial integration. Start with one monitored movie and one episode.

Integration, language profiles and provider setup follow the
[official Bazarr setup guide](https://wiki.bazarr.media/Getting-Started/Setup-Guide/).
Defaults for new items do not configure previously imported items automatically.

## Jellyfin acceptance: same files, readable sidecars

The existing Jellyfin definition mounts `/data/media:/data/media`; manager binds
use the different host prefixes shown above. **Source does not establish that
these locations resolve to the same underlying media.** Before bulk searches,
Luis must confirm each selected Radarr/Sonarr file is also the actual Jellyfin
library file. This task does not invent host aliases or change existing mounts.
If they are separate copies or unrelated roots, stop here and resolve the path
contract with Luis before claiming end-to-end integration.

For a video `Example.mkv`, expected sidecars include `Example.es.srt` and
`Example.en.srt`, adjacent to that video (other supported subtitle formats may
also be returned). Jellyfin recognizes language suffixes on external subtitles;
see its [external-track naming rules](https://jellyfin.org/docs/general/server/media/movies/#external-subtitles-and-audio-tracks).
This integration uses shared files; no Jellyfin API key, plugin or webhook is
required. Luis refreshes/scans the affected library after the initial download.

Acceptance is complete only when Luis verifies:

- Bazarr's public egress uses the scraper VPN address. During a controlled tunnel
  outage, public provider traffic fails closed rather than using host egress;
  DNS behavior also follows the intended VPN policy. Keep the firewall enabled.
  Local manager exceptions and UI access may remain available during the outage.
  Check recovery and namespace attachment after a Luis-controlled VPN recreation.

- Both manager connection tests pass and existing movie/series records import
  with accessible paths and the two-language profile.
- A monitored movie and episode each acquire Spanish and English subtitles where
  providers have suitable results; History records successful downloads. Missing
  provider results are not proof of a mount or API failure.
- Sidecars land beside the intended media, Bazarr can write them, and Jellyfin's
  UID/GID can read them. Both language tracks appear and play correctly in Jellyfin.
- Bazarr configuration survives a Luis-controlled redeploy. Existing scraper
  VPN/tool behavior and manager/playback availability remain intact.

Do not have Bazarr and manual tooling overwrite the same sidecar concurrently.
No automatic FFsubsync coupling or additional synchronization step was introduced.
Repository YAML/link checks establish configuration consistency, not these runtime
results or subtitle quality.

Static validation on 2026-10-15 passed: the 16 existing services other than
`gluetun_scrapers` remain unchanged. Gluetun adds only the Bazarr UI publication,
input-port rule and two manager-IP exceptions (plus explanatory comments).
Four YAML units contain 18 services, eight unchanged subnet-valid static addresses
and matching blank variable contracts. Central/local version tables, local
documentation links and `git diff --check` also passed. No protected
configuration or runtime state was read, and no deployment was performed.

## Rollback boundary

Luis can stop/remove Bazarr and restore the prior subtitle Compose revision through
Portainer if acceptance fails, reverting the added Gluetun port/firewall settings
as well. Coordinate recreation with its other namespace consumers.
Retain Bazarr's `/config` for diagnosis/retry. Shared network and media remain in
place. Removing the service does not remove downloaded sidecars; review those
separately rather than deleting library files during rollback.
