---
title: Mirror Upstream Charts to a Private Registry
description: Mirror Helm chart OCI artifacts and container images from upstream sources to a private registry
sidebar_position: 7
---

# Mirror Upstream Charts to a Private Registry

Mirror Helm charts and their container images from upstream sources to a target OCI registry. This is useful for air-gapped environments, private registry consolidation, and ensuring supply chain availability.

## Create a mirror.yaml Config

Create a `mirror.yaml` file that defines where to mirror from and where to mirror to.

### Minimal example

```yaml
target:
  url: registry.example.com/team

upstreams:
  vault:
    type: repo
    repo: https://helm.releases.hashicorp.com
    chart: vault
    versions:
      - "0.28.0"
```

## Source Types

The mirror command supports three upstream source types.

### Helm repository (repo)

Pull charts from a traditional Helm repository that uses an `index.yaml` catalog.

```yaml
upstreams:
  vault:
    type: repo
    repo: https://helm.releases.hashicorp.com
    chart: vault
    versions:
      - "0.28.0"
      - "0.29.0"
```

### OCI registry (oci)

Pull charts that are already published as OCI artifacts in another registry.

```yaml
upstreams:
  keycloak:
    type: oci
    registry: oci://registry-1.docker.io/bitnamicharts
    chart: keycloak
    versions:
      - "24.0.1"
```

### Local directory (local)

Package a chart from a local directory. Paths are resolved relative to the config file location.

```yaml
upstreams:
  internal-app:
    type: local
    path: ./charts/internal-app
    chart: internal-app
    versions:
      - "1.0.0"
```

## Preview with Dry Run

Always preview what will be mirrored before making changes:

```bash
lazyoci mirror --config mirror.yaml --all --dry-run
```

This shows every chart and image that would be pushed or copied, without actually writing to the target registry.

## Mirror All Charts

Mirror every chart and its container images:

```bash
lazyoci mirror --config mirror.yaml --all
```

The command is idempotent. Charts and images that already exist in the target registry are skipped automatically.

## Mirror a Specific Chart

Mirror a single chart by its key in the config:

```bash
lazyoci mirror --config mirror.yaml --chart vault
```

### Override versions

Override the version list from the config file:

```bash
lazyoci mirror --config mirror.yaml --chart vault --version 0.28.0 --version 0.29.0
```

## Selective Mirroring

### Charts only

Mirror chart OCI artifacts without copying container images:

```bash
lazyoci mirror --config mirror.yaml --all --charts-only
```

### Images only

Copy container images without pushing chart artifacts:

```bash
lazyoci mirror --config mirror.yaml --chart vault --images-only
```

## Target Configuration

### Charts prefix

Group chart artifacts under a path prefix in the target registry:

```yaml
target:
  url: registry.example.com/team
  charts-prefix: charts
```

This pushes charts to `registry.example.com/team/charts/<name>:<version>` instead of `registry.example.com/team/<name>:<version>`.

### Insecure registries

Allow plain HTTP connections to the target registry:

```yaml
target:
  url: localhost:5050
  insecure: true
```

## Image Remapping

Container images extracted from chart templates are remapped to the target registry. The source registry host is stripped and the repository path is preserved:

| Source | Target (url: `registry.example.com/team`) |
|--------|------|
| `ghcr.io/hashicorp/vault:1.15.0` | `registry.example.com/team/hashicorp/vault:1.15.0` |
| `docker.io/library/redis:7.2` | `registry.example.com/team/library/redis:7.2` |
| `quay.io/jetstack/cert-manager:v1.16.1` | `registry.example.com/team/jetstack/cert-manager:v1.16.1` |

## Concurrency

Image copies run in parallel. Adjust the concurrency limit (default 4):

```bash
lazyoci mirror --config mirror.yaml --all --concurrency 8
```

## JSON Output for CI/CD

Use structured output for scripting and pipeline integration:

```bash
lazyoci mirror --config mirror.yaml --all -o json
```

Parse results with `jq`:

```bash
# Check if any images failed
lazyoci mirror --config mirror.yaml --all -o json | jq '.imagesFailed'

# List all copied images
lazyoci mirror --config mirror.yaml --all -o json | \
  jq '[.charts[].versions[].images[] | select(.status == "copied") | .target]'
```

## Troubleshooting

### helm: command not found

The mirror command requires the `helm` CLI to be installed and on your `PATH`. Install it from [helm.sh](https://helm.sh/docs/intro/install/).

### Authentication errors

Credentials are resolved per-registry through the standard lazyoci credential chain. Ensure you are authenticated to both the source and target registries:

```bash
# Authenticate to target
docker login registry.example.com

# Source registries using Docker Hub
docker login docker.io
```

### Image extraction finds no images

Image references are extracted by running `helm template` with default values. If a chart uses non-standard image references or requires custom values to render, some images may not be discovered. You can verify what images are found with a dry run:

```bash
lazyoci mirror --config mirror.yaml --chart vault --dry-run
```

## See Also

- [mirror CLI reference](/reference/cli/mirror) -- full flag and output documentation
- [Authentication guide](/guides/authentication) -- credential chain details
- [Insecure registries](/guides/insecure-registries) -- HTTP registry setup
