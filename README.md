# Media Stack

`dessun-lab/media-stack` owns Luis's media service definitions and operational
tools: video/music streaming, library automation and requests, VPN-backed media
acquisition, subtitle workflows and manually built tool images.

Production repository authority is split into four independent units:
[streaming](stacks/streaming/README.md), [management](stacks/management/README.md),
[acquisition](stacks/acquisition/README.md), [subtitles](stacks/subtitles/README.md).
They retain the shared `arrnet_managed` network and two distinct VPN namespaces.
There is no root aggregate Compose file.

**Luis operates production through Portainer connected directly to GitHub.**
Each unit includes compact PORTAINER.md metadata and a blank `.env.example` for
his configuration. **Cutover is complete:** Luis confirms the old monolithic stack
was removed, all four production stacks are operational, the external network was
provisioned and the Portainer variable contracts were populated. This is operator
evidence, recorded 2026-10-11; agents did not inspect runtime state.
GitHub Actions separately synchronizes a server checkout for host workflows and
manual builds; Portainer does not deploy from that checkout.

## Prerequisite: shared network

Before deploying any production stack, Luis must ensure the USER-owned external
network `arrnet_managed` exists on the target Docker host. If absent, provision it
once:

```bash
docker network create --subnet 172.18.0.0/16 arrnet_managed
```

Reuse the existing network; preserve it across stack removal. See
[shared networking](docs/operations/networking.md) for evidence, static addresses
and ongoing lifecycle responsibilities.

## Find your way

- [Documentation map](docs/README.md): current architecture and operations.
- [Version inventory](docs/versions.md): declared tags, fleet evidence and uncertainty.
- [Bazarr setup](docs/operations/bazarr.md): Spanish/English subtitles through
  Radarr/Sonarr; new repository addition pending Luis's deployment and acceptance.
- [Tools](docs/tools/README.md): locally owned utilities and manual build notes.
- [Canaries](canaries/README.md): temporary migrations; Jellyfin is INACTIVE.
- [Implementation record](docs/design/restructuring.md) and
  [completed migration record](docs/design/implementation-handoff.md).
- [CONTRIBUTING.md](CONTRIBUTING.md): contribution and maintenance rules.
- [AGENTS.md](AGENTS.md): agent authority, effect limits and protected files.

Local Compose use is secondary operator support, not a portable development
setup: definitions retain server-specific binds, names, devices and addresses.
See [deployment](docs/operations/deployment.md). Custom images remain manually
built by Luis. Agents never read or implicitly load the protected root `.env`.

Originally based on [Tom Spark's ARR Stack](https://github.com/loponai/arrstack).
