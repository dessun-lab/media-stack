# Shared production networking

[Documentation map](../README.md) · [Deployment](deployment.md) ·
[Network/variable design](../design/network-and-variables.md)

## Ownership and deployment prerequisite

`arrnet_managed` is persistent shared host infrastructure for the four Media Stack
production projects. It preserves the common network identity and existing static
addresses while services have independent stack lifecycles. **Luis / USER owns
runtime creation, verification, preservation and eventual deletion.** No application
stack owns this network or should create/delete it as part of its lifecycle.

All four Compose files deliberately use the same declaration:

```yaml
networks:
  arrnetwork:
    external: true
    name: arrnet_managed
```

The engine network **must exist before any consuming stack is deployed**. Compose
resolves the literal name; it does not create an external network and fails if it
is absent. See the [Compose network reference](https://docs.docker.com/reference/compose-file/networks/#external).

| Consuming production stack | Authoritative declaration |
| --- | --- |
| Streaming | [compose.yml](../../stacks/streaming/compose.yml) |
| Management | [compose.yml](../../stacks/management/compose.yml) |
| Acquisition | [compose.yml](../../stacks/acquisition/compose.yml) |
| Subtitles | [compose.yml](../../stacks/subtitles/compose.yml) |

## Provision once, only if absent

During runtime cutover, Luis verifies the existing network on the intended Docker
host and reuses it when compatible. If absent, this is the minimal creation command
that reproduces the repository-declared contract:

```bash
docker network create --subnet 172.18.0.0/16 arrnet_managed
```

This is an operator command, not an agent validation step. Do not delete/recreate
an existing network merely to apply this documentation. If its actual configuration
conflicts with the contract, Luis resolves that before deployment.

### Parameter evidence

The pre-split `docker-compose.yml` at commit
`5956f4cbaccebe8d03b1d23eead73aaf98d984f4` declared:

```yaml
networks:
  default:
    name: arrnet_managed
  arrnetwork:
    name: arrnet_managed
    ipam:
      config:
        - subnet: 172.18.0.0/16
```

| Property | Reproduction rule and evidence |
| --- | --- |
| Engine identity | `arrnet_managed`, preserved in all four current declarations. |
| IPv4 subnet | `172.18.0.0/16`, explicitly declared in historical IPAM. |
| Driver | No historical override. Docker's `network create` default is `bridge`; the command relies on it. |
| Gateway | Not explicitly declared historically; omit `--gateway` and let Docker select it. No fixed gateway address is asserted. |
| Other options | No historical IPAM driver, allocation subrange, auxiliary addresses, driver options, internal mode or IPv6 override; no extra flags are required by the source contract. |

Docker documents the [bridge default and gateway selection](https://docs.docker.com/reference/cli/docker/network/create/).
These are source declarations and documented defaults, not a live inspection of
this host. The unused historical `default` alias was omitted during the split;
it does not require a second network. IPAM belongs in host provisioning, not in
the four external Compose declarations.

## Static-address contract

Retain these unique addresses from the pre-split service definitions. All eight
are valid host addresses within `172.18.0.0/16`; current Compose files own the exact
service assignments.

| Existing static address | Service | Owner |
| --- | --- | --- |
| `172.18.0.2` | `jellyfin` | streaming |
| `172.18.0.3` | `gluetun` | acquisition |
| `172.18.0.4` | `radarr` | management |
| `172.18.0.5` | `sonarr` | management |
| `172.18.0.7` | `jellyseerr` | management |
| `172.18.0.8` | `navidrome` | streaming |
| `172.18.0.9` | `lidarr` | management |
| `172.18.0.10` | `gluetun_scrapers` | subtitles |

VPN consumers use their owning Gluetun service's network namespace and therefore
its interface/address; they do not need additional static addresses. Services with
`network_mode: none` remain unattached. Shared network membership does not establish
startup order or make stack lifecycles dependent. The inactive Jellyfin canary's
separate network is outside this production provisioning contract.

## Persistence and cutover

Stopping/removing an individual production stack must leave the shared network
available, including when all four stacks are stopped or removed. Compose does
not remove external networks during `down`; see [Docker's removal contract](https://docs.docker.com/reference/cli/docker/compose/down/).
Keep every production declaration external rather than making one stack the owner.

Luis must preserve the network when retiring the old monolithic stack: its old
Compose declared a managed network, and the new files do not retroactively change
that old project's ownership. Transfer fixed names, addresses and writable state
once, without concurrent old/new owners. See the [cutover handoff](../design/implementation-handoff.md).

External ownership is not protection against host-wide cleanup: unused networks
can be removed by [Docker network pruning](https://docs.docker.com/reference/cli/docker/network/prune/).
Luis's host-maintenance policy must preserve this infrastructure even when no
containers are attached. Eventual deletion is a separate USER decision after all
consumers and future deployment needs are retired, never routine stack cleanup.

This documentation correction performs no network creation, runtime inspection,
Portainer operation or deployment. Luis verifies the actual network at cutover.
