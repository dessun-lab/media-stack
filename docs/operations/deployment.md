# Deployment ownership and local support

[Documentation map](../README.md) · [Cutover handoff](../design/implementation-handoff.md)

## Portainer — Luis only

Portainer reads `dessun-lab/media-stack` directly with its own Git credentials;
it does not deploy from the synchronized server checkout. Luis's last confirmed
production deployment was one stack. Its observed redeploy blast radius motivated
the four independent source units now implemented. No runtime transition was made.

Luis alone creates/deletes stacks, changes Git/Compose settings, supplies variables
and secrets, chooses update policies, deploys, cuts over and rolls back. Akane and
Avlora never access or inspect Portainer. Its settings are not a repository-work gate.
Each stack's PORTAINER.md provides the recommended name, exact Git Compose path
and blank variable contract. No credential/URL or active environment is documented.

| Action | Effect and owner |
| --- | --- |
| Local commit | Records reviewed source; authorized separately from push or deployment. |
| Publish main | Triggers separate SSH synchronization. No push occurred in implementation. |
| Manual custom-image build | Luis produces a tool image/tag; this is not deployment. |
| Portainer update/cutover | Luis selects units, transfers ownership and validates runtime. |

## Shared network prerequisite

All four production units require the externally managed, USER-owned
`arrnet_managed` before deployment. Luis provisions it once if absent and preserves
it through stack removal. The [networking authority](networking.md) supplies the
creation command, historical subnet evidence and static-address contract.

## Secondary local use

There is no root Compose startup path. A prepared operator selects one unit, for
example `docker compose -f stacks/streaming/compose.yml up -d`. This is an operator
example, not an executed validation command. It requires configured values,
existing images/network, compatible host paths/devices and a deliberately chosen
non-conflicting Docker endpoint. Fixed names/IPs/binds mean changing project name
alone does not isolate a local run. Blank examples are variable lists, not live config.

Agents do not execute Compose in the checkout: it may load protected `.env`.
Use plain parsing for structural checks; never execute the subtitle script as a
test because it launches containers and writes media.

## Runtime handoff

Luis prepares the new infrastructure checkout's server-only music/cookie sources,
existing shared network and manual image availability. Transfer fixed names,
addresses and writable state once; prevent old-project retirement from removing
the shared network. Luis owns interruption, backups, rollback and runtime checks.
A file deletion does not stop an existing container; this includes deunhealth.
The [Jellyfin canary](../../canaries/jellyfin/README.md) remains inactive and unsafe
for concurrent activation until deliberately isolated.
