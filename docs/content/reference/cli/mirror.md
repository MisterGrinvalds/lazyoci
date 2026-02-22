---
title: mirror
---

# mirror

Mirror Helm chart OCI artifacts and their referenced container images from upstream sources to a target OCI registry.

## Synopsis

```
lazyoci mirror [flags]
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `mirror.yaml` | Path to mirror configuration YAML file |
| `--chart` | | `""` | Mirror a specific chart by its key in the config |
| `--all` | | `false` | Mirror all charts defined in the config |
| `--version` | | `[]` | Override version(s) to mirror (repeatable) |
| `--dry-run` | | `false` | Preview what would be mirrored without making changes |
| `--charts-only` | | `false` | Mirror chart OCI artifacts only, skip container images |
| `--images-only` | | `false` | Mirror container images only, skip chart push |
| `--concurrency` | | `4` | Number of parallel image copies per chart version |

**Mutual exclusivity:** `--chart` and `--all` cannot be used together; exactly one must be specified.

**Repeatable flag:** `--version` can be specified multiple times to override the config file's version list (e.g. `--version 0.28.0 --version 0.29.0`).

## Inherited Flags

| Flag | Short | Default | Values |
|------|-------|---------|--------|
| `--output` | `-o` | `text` | `text`, `json`, `yaml` |
| `--artifact-dir` | | `""` | Artifact storage directory |
| `--theme` | | `""` | Theme name |

## Configuration

The mirror command reads a YAML configuration file (`--config`) that defines:

- **Target registry**: destination URL, insecure mode, and optional charts prefix
- **Upstreams**: a map of chart keys to their source definitions

Three upstream source types are supported:

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `repo` | Traditional Helm repo (index.yaml) | `repo`, `chart`, `versions` |
| `oci` | OCI registry with chart artifacts | `registry`, `chart`, `versions` |
| `local` | Chart directory on local filesystem | `path`, `chart`, `versions` |

### Config schema

```yaml
target:
  url: <registry-host>/<optional-path>   # required
  insecure: false                         # allow plain HTTP
  charts-prefix: ""                       # path segment between URL and chart name

upstreams:
  <key>:                                  # arbitrary identifier
    type: repo | oci | local              # required
    repo: <helm-repo-url>                 # type=repo only
    registry: oci://<registry-url>        # type=oci only
    path: ./relative/or/absolute          # type=local only (resolved relative to config file)
    chart: <chart-name>                   # required
    versions:                             # explicit version list
      - "0.28.0"
```

### Example `mirror.yaml`

```yaml
target:
  url: registry.example.com/ns
  insecure: false
  charts-prefix: charts

upstreams:
  vault:
    type: repo
    repo: https://helm.releases.hashicorp.com
    chart: vault
    versions:
      - "0.28.0"
      - "0.29.0"

  keycloak:
    type: oci
    registry: oci://registry-1.docker.io/bitnamicharts
    chart: keycloak
    versions:
      - "24.0.1"

  demo-app:
    type: local
    path: ./charts/demo-app
    chart: demo-app
    versions:
      - "1.0.0"
```

## How It Works

1. For each chart + version pair, the chart is pulled from the upstream source
2. The chart is pushed as an OCI artifact to the target registry (skipped if already present)
3. Container image references are extracted by running `helm template` and parsing `image:` lines
4. Each image is remapped to the target registry (source host stripped, path preserved)
5. Images are copied registry-to-registry in parallel (skipped if already present)

Credentials are resolved per-registry through the standard lazyoci credential chain, ensuring credentials never leak between source and target registries.

### Prerequisites

The `helm` CLI must be installed and available on `PATH`.

## Examples

```bash
# Mirror all charts defined in the config
lazyoci mirror --config mirror.yaml --all

# Mirror a specific chart
lazyoci mirror --config mirror.yaml --chart vault

# Mirror with version override
lazyoci mirror --config mirror.yaml --chart vault --version 0.28.0

# Mirror multiple versions
lazyoci mirror --config mirror.yaml --chart vault --version 0.28.0 --version 0.29.0

# Dry run -- preview what would be mirrored
lazyoci mirror --config mirror.yaml --all --dry-run

# Charts only (skip container images)
lazyoci mirror --config mirror.yaml --chart vault --charts-only

# Images only (skip chart push)
lazyoci mirror --config mirror.yaml --chart vault --images-only

# JSON output for scripting
lazyoci mirror --config mirror.yaml --all -o json
```

## Output

### Text output

```
Mirror: vault (vault) → registry.example.com/ns/charts
Versions: 0.28.0

── vault:0.28.0 ──
  Chart: pushing... OK
  Images: 3 found
    docker.io/library/redis:7.2 → copying...
    docker.io/library/redis:7.2 → OK
    ghcr.io/hashicorp/vault:1.15.0 → exists
    ghcr.io/hashicorp/vault-k8s:1.4.0 → copying...
    ghcr.io/hashicorp/vault-k8s:1.4.0 → OK

════════════════════════════════════════
Mirror Summary
════════════════════════════════════════
  vault (vault):
    0.28.0: chart=pushed  images=2 copied, 1 skipped, 0 failed

  Charts:  1 pushed, 0 skipped, 0 failed
  Images:  2 copied, 1 skipped, 0 failed
```

### JSON output

```json
{
  "charts": [
    {
      "key": "vault",
      "chart": "vault",
      "versions": [
        {
          "version": "0.28.0",
          "chartStatus": "pushed",
          "images": [
            {
              "source": "docker.io/library/redis:7.2",
              "target": "registry.example.com/ns/library/redis:7.2",
              "status": "copied"
            }
          ],
          "imagesCopied": 2,
          "imagesSkipped": 1,
          "imagesFailed": 0
        }
      ]
    }
  ],
  "chartsPushed": 1,
  "chartsSkipped": 0,
  "chartsFailed": 0,
  "imagesCopied": 2,
  "imagesSkipped": 1,
  "imagesFailed": 0
}
```

## See Also

- [Examples: mirror/](https://github.com/mistergrinvalds/lazyoci/tree/main/examples/mirror) -- example mirror configuration with all three source types
