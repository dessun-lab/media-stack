# Media Stack service improvement audit — October 2026

[Documentation map](../README.md) · [Version authority](../versions.md) ·
[Completed migration](../design/implementation-handoff.md)

**2026-10-11 — review only; no recommendation implemented.** Luis confirms the
four production stacks are operational after cutover. This audit inspected source,
not Portainer, Docker, host state or application settings. A missing Compose option
is not proof that a mounted configuration or external proxy lacks that setting.

## Evidence and version limits

Baseline: local Media Stack `162131b`; all four Compose files, local recipes,
NensaySubs source and the subtitle workflow. Read-only fleet evidence is
`homelab-inventory/homelab-inventory.md` at
`32a87585bf6404ed4962d1ebf0945d97210ac26d`, inventoried 2026-10-03. Its 15 matching
service image/tag rows, absent MakeMKV/FFsubsync rows and stale deunhealth row are
recorded in [versions](../versions.md). No fleet edit or version upgrade occurred.

Official sources below were consulted on 2026-10-11. Version-tagged upstream
source is preferred; rolling documentation is identified where used. Jellyfin's
literal `12.1ubu2604-ls50`/fleet `12.1` identity is retained without substituting a
familiar release number. No exact-build Jellyfin feature claim is made. Custom
`latest` images have no established deployed binary/package revision; their
repository recipes are not deployment attestations. Newer documentation alone
cannot establish support in those images.

**RECOMMEND** means an evidence-backed next action; **CONSIDER** needs workload or
operator context; **SKIP** records a useful negative decision. Every implementation
or operator check below requires its own scope. Nothing requests secrets.

## Service coverage

Exact image identities and build suffixes remain in the central version authority;
versions below identify the feature-research baseline, not a second inventory.

| Stack | Service | Relevant version evidence | Focus and disposition |
| --- | --- | --- | --- |
| streaming | `jellyfin` | `12.1ubu2604-ls50`, fleet `12.1` | GPU/device and plugin binds, writable media/canary overlap reviewed. No material new tuning identified; S1. |
| streaming | `navidrome` | `0.64.2` | Scanner/watcher, cache, auth and metrics reviewed. Metrics access: R2. Existing scanning and cache knobs do not establish a performance problem. |
| management | `radarr` | app `6.4.4.10685` | Download/import path and cross-stack integration: R4. No version-specific feature switch justified. |
| management | `sonarr` | app `4.0.20.3014` | Download/import path and hardlink boundaries: R4. |
| management | `lidarr` | `latest`; fleet app `3.1.0.4875` | R4 includes downloads and distinct Navidrome music source. No feature advice inferred from mutable tag. |
| management | `jellyseerr` | Seerr `v3.5.0` | Actual image ownership and process lifecycle: R3; retain historical service name. |
| acquisition | `gluetun` | `v3.41.3` | Built-in health/recovery and namespace lifecycle: C1/S2. |
| acquisition | `qbittorrent` | app `5.2.4`, libtorrent `2.0.15` | Existing VPN health gate retained; import paths R4. Interface binding reviewed, S3. |
| acquisition | `prowlarr` | app `2.6.5.5623` | FlareSolverr integration checked in [tagged proxy source](https://raw.githubusercontent.com/Prowlarr/Prowlarr/v2.6.5.5623/src/NzbDrone.Core/IndexerProxies/FlareSolverr/FlareSolverr.cs); same namespace permits loopback access. Active proxy/app settings were not inspected; no demonstrated integration defect. |
| acquisition | `flaresolverr` | `v3.5.2` | Optional request/outcome metrics: C2. No need to publish solver API merely for Prowlarr's same-namespace access. |
| acquisition | `ytdlp` | custom `latest`, unpinned pip | Manual tool readiness/recreation: C1. YouTube dependency completeness: C3, conditional on actual package version. |
| acquisition | `themearr` | custom `latest`, fleet N/A | Main VPN, CLI port 7681 and output bind reviewed. External source provenance is documented in [tools](../tools/README.md#themearr-external-source); no new feature claim without deployed revision or workload evidence. |
| acquisition | `makemkv` | custom `latest`; recipe default `1.18.3` | Offline CLI, `HOME=/config`, persisted config and download bind fit its role. No material improvement identified; no need to add networking/disc devices for an unrequested workflow. |
| subtitles | `gluetun_scrapers` | `v3.41.3` | Independent namespace and built-in health: C1/S2. Keep separate from acquisition VPN. |
| subtitles | `flaresolverr_subs` | `v3.5.2` | NensaySubs timeout contract R1; optional metrics C2. |
| subtitles | `nensaysubs` | custom `latest`; local Go source | Request budget mismatch R1; solver `service_started` is not readiness, C1. Source/binary parity remains unknown. |
| subtitles | `ffsubsync` | custom `latest`, unpinned pip | Offline `run --rm`, separate `.synced` output and batch behavior reviewed; preserve manual operation, S4. |

## Recommended next actions

### R1 — Align subtitle solver and client time budgets

- **Stack / service:** subtitles / NensaySubs and FlareSolverr Subs.
- **Finding / why:** [local source](../../tools/nensaysubs/nensaysubs.go) sets
  `flareSolverrTimeout=120000`, while `newHTTPClient` sets a 25-second
  `ResponseHeaderTimeout` and 90-second overall timeout. `callFlareSolverr` uses
  that same client. A solve that delays headers beyond 25 seconds can fail on the
  client before the requested solver budget expires; this is a code-level risk,
  not an observed production failure.
- **Version evidence:** local `newHTTPClient`, `solveWithFlareSolverr` and
  `callFlareSolverr`; [FlareSolverr v3.5.2 API](https://raw.githubusercontent.com/FlareSolverr/FlareSolverr/v3.5.2/README.md)
  defines `maxTimeout` in milliseconds. [Go's HTTP contract](https://pkg.go.dev/net/http#Transport)
  defines the header wait separately from the overall client timeout. The deployed
  custom-image revision is unknown.
- **Benefit / tradeoff:** avoid prematurely abandoning legitimate slow solves;
  longer waits can delay failure feedback. Do not relax ordinary site requests.
- **RECOMMEND:** authorize a focused source fix using a solver-specific bounded
  client budget above the solve deadline, plus a local delayed-response test.
  Establish deployed source parity before any separately authorized manual build.

### R2 — Establish access protection for the already enabled metrics endpoint

- **Stack / service:** streaming / Navidrome.
- **Finding / why:** [Compose](../../stacks/streaming/compose.yml) enables
  `ND_PROMETHEUS_ENABLED` at `/metrics`, with no metrics password declared. The
  service publishes 4533. Proxy controls and mounted configuration are unknown;
  this is an access-policy gap in evidence, not proof of public exposure.
- **Version evidence:** [v0.64.2 configuration](https://raw.githubusercontent.com/navidrome/navidrome/v0.64.2/conf/configuration.go)
  supports `Prometheus.Password`, default empty. The [official options](https://navidrome.org/docs/usage/configuration/options/)
  document `ND_PROMETHEUS_PASSWORD` Basic Auth with user `navidrome`.
- **Benefit / tradeoff:** keep useful metrics within the monitoring trust boundary;
  enabling auth requires coordinated scraper credentials and could interrupt scraping.
- **RECOMMEND:** Luis first confirms the intended collector and existing access
  controls. If absent, choose endpoint auth or restricted routing; if unused,
  consider disabling metrics. No credential belongs in repository documentation.

### R3 — Correct Seerr's image-specific identity assumptions; consider init

- **Stack / service:** management / `jellyseerr` (official Seerr image).
- **Finding / why:** Compose passes `PUID`/`PGID` but has no `init`. The
  [v3.5.0 Dockerfile](https://raw.githubusercontent.com/seerr-team/seerr/v3.5.0/Dockerfile)
  runs as `node:node` and starts through `npm`; it has no UID-remapping entrypoint.
  The [maintainer explanation](https://docs.seerr.dev/blog/why-seerr-doesnt-support-puid-pgid/)
  explicitly says those environment variables are unsupported.
- **Version evidence:** exact v3.5.0 packaging above; rolling [official Docker guide](https://docs.seerr.dev/getting-started/docker/)
  includes `init: true`. Init is a container process-management option, not a claim
  that a newer Seerr application release is required.
- **Benefit / tradeoff:** remove misleading identity controls and provide deliberate
  child-process/signal handling; changing process supervision needs a controlled
  service update. No existing zombie/shutdown issue was observed.
- **RECOMMEND:** approve a small packaging follow-up: remove ineffective variables
  from this service only; **CONSIDER** adding init in the same review. Preserve
  working data ownership; do not blindly chown data or change the user. Other
  LinuxServer services still consume PUID/PGID.

### R4 — Verify the import/storage contract before considering path changes

- **Stack / service:** management plus acquisition/streaming: Radarr, Sonarr,
  Lidarr, qBittorrent and Navidrome.
- **Finding / why:** qBittorrent binds `/data/fast/qbittorrent` to `/data`, whereas
  managers bind separate `runtime/arrstack/<manager>/downloads` trees. Libraries
  use additional mounts. Lidarr music and Navidrome music also have distinct host
  sources. Source alone cannot establish host aliases, shared filesystems, remote
  path mappings or an intended automatic Lidarr-to-Navidrome pipeline.
- **Version evidence:** the current Compose mounts; [official Sonarr Docker guidance](https://wiki.servarr.com/en/sonarr/installation/docker)
  explains that inconsistent paths and separate container mounts impede imports,
  hardlinks and fast moves. This filesystem constraint is not a new-release knob.
- **Benefit / tradeoff:** identify avoidable copies or misunderstood handoff paths;
  any real storage migration risks paths, seeding and data access. Successful
  cutover is not evidence that an import or hardlink acceptance test occurred.
- **RECOMMEND:** a bounded Luis-led review of one existing completed import per
  manager and the intended music handoff. Keep working mappings if deliberate;
  propose a separate storage change only after evidence of a concrete benefit.

## Conditional opportunities

### C1 — Define recovery expectations without adding another watchdog

- **Stack / service:** acquisition and subtitles / both Gluetuns and namespace
  consumers, especially yt-dlp and NensaySubs.
- **Finding / why:** yt-dlp is health-gated but lacks the dependency `restart: true`
  used by most peers. NensaySubs waits only for solver `service_started`, although
  browser readiness can lag process start. Its manual invocation may already
  avoid the race; no outage was reported.
- **Version evidence:** Gluetun [v3.41.3 image healthcheck](https://raw.githubusercontent.com/qdm12/gluetun/v3.41.3/Dockerfile)
  and [health defaults](https://raw.githubusercontent.com/qdm12/gluetun/v3.41.3/internal/configuration/settings/health.go)
  provide VPN auto-healing by default. [Compose startup semantics](https://docs.docker.com/compose/how-tos/startup-order/)
  distinguish started/healthy and explicit dependency restarts. They do not prove
  a particular Portainer recovery outcome or continuous supervision.
- **Benefit / tradeoff:** distinguish internal tunnel recovery, VPN-container
  replacement and solver startup; forced restarts can interrupt manual work.
- **CONSIDER:** Luis records expected behavior at the next planned maintenance.
  If failures exist, scope yt-dlp reattachment/recreation and solver readiness
  separately. No daemon, automatic restarter or provider-specific firewall change.

### C2 — Add solver metrics only when there is a monitoring consumer

- **Stack / service:** acquisition/subtitles / both FlareSolverr services.
- **Finding / version evidence:** Compose has no exporter settings.
  [v3.5.2 documentation](https://raw.githubusercontent.com/FlareSolverr/FlareSolverr/v3.5.2/README.md)
  provides `PROMETHEUS_ENABLED` and `PROMETHEUS_PORT` (default 8192) with request
  outcome and duration metrics.
- **Benefit / tradeoff:** distinguish slow/failed solving from VPN failures;
  additional scrape routing and access controls need ownership. Each solver shares
  its own VPN owner's interface; the same port can serve the two separate namespaces.
- **CONSIDER:** only if Luis wants these signals in existing monitoring. Select a
  reachable restricted target on each VPN owner; do not publish ports on namespace
  consumers or expose metrics solely to make collection convenient.

### C3 — Check yt-dlp's YouTube dependencies at its next manual build review

- **Stack / service:** acquisition / yt-dlp.
- **Finding / evidence:** [recipe](../../tools/yt-dlp/Dockerfile) installs unpinned
  `yt-dlp` with pip and FFmpeg, but declares no JavaScript runtime or EJS dependency
  group. The current [official EJS guide](https://github.com/yt-dlp/yt-dlp/wiki/EJS)
  requires a supported runtime and matching solver scripts for YouTube; the pip
  `default` group supplies the latter, not the runtime.
- **Version limit:** `latest`/fleet N/A does not establish the installed yt-dlp
  version. Rolling requirements are conditional until that version and use of
  YouTube are established; no claim that the running tool is broken.
- **Benefit / tradeoff:** a future manual rebuild can explicitly include the
  dependencies its chosen version needs; extra dependencies require maintenance.
- **CONSIDER:** record the package version and intended extractors, then evaluate
  dependencies for that version. No automatic updater, rebuild or cookie access.

## Useful no-change decisions

- **S1 — SKIP generic Jellyfin GPU/cache tuning.** GPU devices/groups, plugin DLLs
  and cache/transcode binds are already present. [Official acceleration guidance](https://jellyfin.org/docs/general/post-install/transcoding/hardware-acceleration/)
  is hardware-dependent. No GPU model, load evidence or exact-image feature proof
  justifies enabling guessed flags, relocating cache or adding privileged access.
  Keep the inactive canary isolated from production lifecycle.
- **S2 — SKIP restoring deunhealth or adding duplicate Gluetun healthchecks.**
  Both declared v3.41.3 images already supply healthchecks and internal VPN recovery
  (C1). That does not heal every application failure, but no reported failure
  justifies restoring Docker-socket restart authority removed by design.
- **S3 — SKIP guessed qBittorrent interface/port-forwarding settings.**
  [5.2.4 source](https://raw.githubusercontent.com/qbittorrent/qBittorrent/release-5.2.4/src/base/bittorrent/sessionimpl.cpp)
  supports interface selection, but the active setting, VPN interface and provider
  capabilities are outside this audit. Preserve Gluetun ownership; do not invent
  a `tun0`/WireGuard binding or assume forwarding support.
- **S4 — SKIP bulk FFsubsync flag changes without package/workload evidence.**
  [Upstream](https://github.com/smacke/ffsubsync) describes alternative alignment
  and quality controls; support in the unpinned local image is unestablished.
  The existing [workflow](../../scripts/download-and-sync-subs.bash) keeps originals
  and writes separate `.synced` outputs. No demonstrated alignment failure warrants
  global flags, automatic replacement or an always-running alignment worker.

## Stack-level disposition

| Operational unit | Assessment |
| --- | --- |
| Streaming | Independent playback lifecycle is working per Luis. Existing scanner/cache settings need no shared tuning layer; focus R2 and keep Jellyfin hardware settings workload-driven. Shared host/media remain cross-unit dependencies. |
| Management | Independent interfaces and application-level integrations are sensible; no startup barrier on acquisition is needed for the grouping. R3 addresses packaging; R4 evaluates imports without moving ownership. Avoid shared environment templates that assume every image implements LinuxServer UID behavior. |
| Acquisition | VPN consumers remain with their namespace owner; port ownership is correct. C1/R4 address recovery/data handoffs; C2/C3 are conditional. Manual tools do not justify another permanent stack or build automation. |
| Subtitles | Separate VPN/solver loopback and offline alignment fit the workflow. R1 is a concrete source-level reliability candidate; C1 distinguishes browser readiness from container start. No reason to merge VPN namespaces or convert tools to new lifecycle models. |

No substantial defect was found in the approved four-stack boundaries. This was
not an exhaustive security assessment, performance benchmark or runtime test.
Research stopped at concrete opportunities and version limits; release freshness
alone is not a recommendation.

## Repository validation

Static validation confirms all five Compose documents parse identically to HEAD,
including service order, with every non-comment line unchanged. All 17 production
services and the inactive canary have separator/role headers. External network
identity, eight static addresses and all version declarations are unchanged.
All 17 services appear in this review; 15 fleet image/tag rows match and the two
missing tool rows remain explicit. Local Markdown links/anchors, fences and scoped
`git diff --check` pass. No recommendation, runtime action or secret access occurred.

## Review boundary

Luis/Avlora should select one finding for a separate bounded brief; R1 is the
strongest source-only starting point. A future brief must distinguish source/tests,
manual image build and Luis-owned deployment. The applied restructuring remains
complete regardless of which optional improvements are selected.
