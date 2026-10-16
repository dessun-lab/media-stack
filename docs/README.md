# Documentation map

Media Stack documents repository declarations separately from deployed state.
Luis confirms the four production stacks are operational after successful cutover
(operator update recorded 2026-10-11). The old monolith was removed; the shared
network and Portainer variables are provisioned. Runtime authority remains Luis's.
Luis confirms Bazarr is deployed and functioning in Media Stack (operator update
recorded 2026-10-15). Its repository configuration and operation are documented
under [Bazarr](operations/bazarr.md).

| Subject | Owning authority |
| --- | --- |
| Architecture and shared contracts | [Architecture](architecture.md) |
| Unit membership, service details and declared versions | [Streaming](../stacks/streaming/README.md), [management](../stacks/management/README.md), [acquisition](../stacks/acquisition/README.md), [subtitles](../stacks/subtitles/README.md) |
| Exact service definitions | Each unit's `compose.yml` |
| Portainer name/path metadata and variable contracts | Each unit's `PORTAINER.md` and blank `.env.example` |
| Version provenance, fleet comparison and discrepancies | [Versions](versions.md) |
| Portainer authority and secondary local operation | [Deployment](operations/deployment.md) |
| Shared network provisioning, static addresses and lifecycle | [Networking](operations/networking.md) |
| Actions → SSH → server checkout | [Synchronization](operations/synchronization.md) |
| Cross-tool subtitle workflow | [Subtitle operation](operations/subtitles.md) |
| Bazarr configuration and operation | [Bazarr](operations/bazarr.md) |
| Tool source and manual builds | [Tool index](tools/README.md), then tool-local READMEs |
| Temporary migration operation/history | [Canaries](../canaries/README.md), then scenario README/MIGRATIONS |
| Approved structure, rationale and acceptance | [Restructuring](design/restructuring.md), [unit decisions](design/production-units.md), [network/variables](design/network-and-variables.md) |
| Completed cutover and path-change history | [Migration record](design/implementation-handoff.md) |
| Proposed service improvements (not implemented) | [October audit](audits/service-improvements-2026-10.md) |
| Evidence, permitted differences and validation | [Evidence record](evidence.md) |

[CONTRIBUTING](../CONTRIBUTING.md) governs additions/maintenance;
[AGENTS](../AGENTS.md) routes agents and protects configuration.

Avoid competing descriptions: Compose/tool source owns declared configuration;
local READMEs summarize use and tags; versions.md owns provenance/comparison;
operations pages own workflows across components. The former grouped service
manuals were consolidated into stack owners. Design/history records must not
claim deployment or silently override current source/operator evidence.
