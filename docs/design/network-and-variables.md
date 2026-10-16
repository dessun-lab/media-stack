# Shared network and variable contracts

[Design overview](restructuring.md) · [Production units](production-units.md)

**Approved implementation contract.** Source is the four `stacks/*/compose.yml`
files; no protected `.env` input is needed.

## Preserve the existing network

Before the split, `default` and `arrnetwork` both named `arrnet_managed`; `arrnetwork`
declared `172.18.0.0/16` IPAM and neither entry was external. Every retained service
uses `arrnetwork`, a VPN service namespace or `none`. No retained service consumed
the implicit default network.

Each production Compose file uses this top-level declaration:

```yaml
networks:
  arrnetwork:
    external: true
    name: arrnet_managed
```

Keep existing service-level network entries and namespace references. The split omits the
unused `default` entry and project-owned IPAM declaration. With `external: true`,
Compose references an existing network; it does not create it, and other attributes
besides `name` are invalid. This transfers lifecycle responsibility outside the
projects, not addressing or connectivity.
[Docker network reference](https://docs.docker.com/reference/compose-file/networks/#external).

The operational authority is [shared networking](../operations/networking.md):
it owns the provisioning command, historical parameters, static-address inventory
and Luis's lifecycle/cutover responsibilities. The restructuring introduced no new
network, subnet, DNS alias, static address or VPN identity. The subsequent Bazarr
addition shares `gluetun_scrapers` at `172.18.0.10`, with no additional address.
Namespace consumers retain their owner's interface; `none` services remain unattached. The disabled canary is
outside this production contract.

## One blank variable template per unit

Each production unit contains `stacks/<unit>/.env.example` with exactly the unique
interpolation names from that unit's active Compose values. Every assignment is
`NAME=`; no operational value, secret, synthetic value or copied default. Templates
are documentation for Luis to populate in Portainer, not automatically loaded
configuration. Do not add `env_file:` or commit real `.env` files.

The authoritative name lists live beside each Compose:
[streaming](../../stacks/streaming/.env.example),
[management](../../stacks/management/.env.example),
[acquisition](../../stacks/acquisition/.env.example),
[subtitles](../../stacks/subtitles/.env.example).

These are interpolation inputs, not all container environment keys. Literal
`NENSAYSUBS_*`, `HOME` and Navidrome options stay in service definitions and do not
become inputs. MakeMKV/FFsubsync add no interpolation variables.

| Existing defaulted input | Default remains in Compose |
| --- | --- |
| `GID_VIDEO` | `44` |
| `GID_RENDER` | `993` |
| `ND_EXTAUTH_TRUSTEDSOURCES` | `10.10.0.1/32` |
| `QBITTORRENT_WEBUI_PORT_INTERNAL` | `8080` |
| `ANIMETHEMES_API_URL` | `http://api.animethemes.moe` |

Retain the original `${NAME:-default}` expressions. They use the default when
unset or empty. Other inputs have no declared defaults; the contract marks them
for Luis to supply, without changing expressions into new `:?` checks.
[Docker interpolation reference](https://docs.docker.com/reference/compose-file/interpolation/).

Generate name sets by parsing only selected Compose YAML and scanning scalar
values for interpolation; deduplicate and compare against each example. Never run
Compose against the checkout, inspect `.env`, infer values or copy active settings.
With Bazarr, the production union has 27 unique inputs (36 entries across the four
contracts). Subtitles adds `BAZARR_HOST`, `PUID` and `PGID`; the latter two already
exist in other units. `BAZARR_HOST` is the UI's host bind IP, following the managers'
host-port convention; Luis must explicitly supply it in Portainer.
Validation uses plain YAML parsing and blank examples without interpolation.
These blank templates alone are not deployable configuration.
