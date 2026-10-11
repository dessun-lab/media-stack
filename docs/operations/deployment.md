# Deployment ownership and local support

[Documentation map](../README.md) · [Completed migration](../design/implementation-handoff.md)

## Portainer — Luis only

Portainer reads `dessun-lab/media-stack` directly with its own Git credentials;
it does not deploy from the synchronized server checkout. Luis confirms all four
production stacks are operational after removing the monolithic stack and
populating the network/variable contracts (update recorded 2026-10-11). The old
stack's redeploy blast radius motivated these independent operational units.

Luis alone creates/deletes stacks, changes Git/Compose settings, supplies variables
and secrets, chooses update policies, deploys, cuts over and rolls back. Akane and
Avlora never access or inspect Portainer. Its settings are not a repository-work gate.
Each stack's PORTAINER.md provides the recommended name, exact Git Compose path
and blank variable contract. No credential/URL or active environment is documented.

| Action | Effect and owner |
| --- | --- |
| Local commit | Records reviewed source; authorized separately from push or deployment. |
| Publish main | Triggers separate SSH synchronization; each publication needs authorization. |
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

## Ongoing runtime ownership

The restructuring cutover is complete; [the migration record](../design/implementation-handoff.md)
retains the applied path/ownership changes. Luis continues to own server-only
music/cookie sources, manual image availability, backups, interruption, rollback
and runtime validation. Preserve fixed names, addresses and writable-state ownership
when maintaining stacks, and retain the external network through stack removal.
The [Jellyfin canary](../../canaries/jellyfin/README.md) remains inactive and unsafe
for concurrent activation until deliberately isolated. Its status is separate from
the completed production restructuring.
