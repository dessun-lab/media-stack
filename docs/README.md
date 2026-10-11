# Documentation map

Media Stack documents repository declarations separately from deployed state.
Luis's last confirmed deployment was one Portainer stack; the four-unit source
layout is implemented here, with publication and runtime cutover still Luis-owned.

| Subject | Owning authority |
| --- | --- |
| Architecture and shared contracts | [Architecture](architecture.md) |
| Unit membership, service details and declared versions | [Streaming](../stacks/streaming/README.md), [management](../stacks/management/README.md), [acquisition](../stacks/acquisition/README.md), [subtitles](../stacks/subtitles/README.md) |
| Exact service definitions | Each unit's `compose.yml` |
| Portainer name/path metadata and variable contracts | Each unit's `PORTAINER.md` and blank `.env.example` |
| Version provenance, fleet comparison and discrepancies | [Versions](versions.md) |
| Portainer authority and secondary local operation | [Deployment](operations/deployment.md) |
| Actions → SSH → server checkout | [Synchronization](operations/synchronization.md) |
| Cross-tool subtitle workflow | [Subtitle operation](operations/subtitles.md) |
| Tool source and manual builds | [Tool index](tools/README.md), then tool-local READMEs |
| Temporary migration operation/history | [Canaries](../canaries/README.md), then scenario README/MIGRATIONS |
| Approved structure, rationale and acceptance | [Restructuring](design/restructuring.md), [unit decisions](design/production-units.md), [network/variables](design/network-and-variables.md) |
| Publication, path migration and Luis's cutover | [Implementation handoff](design/implementation-handoff.md) |
| Evidence, permitted differences and validation | [Evidence record](evidence.md) |

[CONTRIBUTING](../CONTRIBUTING.md) governs additions/maintenance;
[AGENTS](../AGENTS.md) routes agents and protects configuration.

Avoid competing descriptions: Compose/tool source owns declared configuration;
local READMEs summarize use and tags; versions.md owns provenance/comparison;
operations pages own workflows across components. The former grouped service
manuals were consolidated into stack owners. Design/history records must not
claim deployment or silently override current source/operator evidence.
