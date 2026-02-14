---
title: Build and Push OCI Artifacts
description: Build container images, Helm charts, and generic artifacts from a .lazy config
sidebar_position: 4
---

# Build and Push OCI Artifacts

Build and push OCI artifacts to registries using a `.lazy` configuration file.

## Create a .lazy Config

Create a `.lazy` file in your project root. This YAML file defines what to build and where to push.

### Minimal example

```yaml
version: 1
artifacts:
  - type: image
    name: myapp
    targets:
      - registry: ghcr.io/owner/myapp
        tags:
          - latest
```

## Build a Container Image

Build a container image from a Dockerfile using `docker buildx`.

### Basic image build

```yaml
version: 1
artifacts:
  - type: image
    name: myapp
    dockerfile: Dockerfile
    context: "."
    targets:
      - registry: ghcr.io/owner/myapp
        tags:
          - "{{ .Tag }}"
          - latest
```

```bash
lazyoci build --tag v1.0.0
```

### Multi-architecture build

```yaml
- type: image
  name: myapp
  platforms:
    - linux/amd64
    - linux/arm64
  targets:
    - registry: ghcr.io/owner/myapp
      tags:
        - "{{ .Tag }}"
```

### With build arguments

```yaml
- type: image
  name: myapp
  buildArgs:
    GO_VERSION: "1.23"
    APP_ENV: production
  targets:
    - registry: ghcr.io/owner/myapp
      tags:
        - "{{ .Tag }}"
```

## Package a Helm Chart

Package a Helm chart directory and push it as an OCI artifact. No Helm CLI needed.

### Helm chart build

```yaml
version: 1
artifacts:
  - type: helm
    name: mychart
    chartPath: charts/mychart
    targets:
      - registry: ghcr.io/owner/charts/mychart
        tags:
          - "{{ .ChartVersion }}"
          - latest
```

```bash
lazyoci build --tag v1.0.0
```

The `{{ .ChartVersion }}` variable is automatically read from `Chart.yaml`.

### Pull with Helm after pushing

```bash
# After pushing with lazyoci
helm pull oci://ghcr.io/owner/charts/mychart --version 0.1.0
```

## Push a Generic Artifact

Push arbitrary files with custom media types. Useful for configs, policies, WASM modules, or documentation.

### Config and policy files

```yaml
version: 1
artifacts:
  - type: artifact
    name: app-config
    mediaType: application/vnd.example.config.v1
    files:
      - path: config.json
        mediaType: application/vnd.example.config.v1+json
      - path: policy.rego
        mediaType: application/vnd.openpolicyagent.policy.layer.v1+rego
    targets:
      - registry: ghcr.io/owner/app-config
        tags:
          - "{{ .Tag }}"
```

```bash
lazyoci build --tag v1.0.0
```

### Pull with ORAS after pushing

```bash
oras pull ghcr.io/owner/app-config:v1.0.0
```

## Push a Docker Daemon Image

Push an image that already exists in your local Docker daemon to a registry. No build step needed.

### Export and push

```yaml
version: 1
artifacts:
  - type: docker
    name: local-app
    image: myapp:latest
    targets:
      - registry: ghcr.io/owner/myapp
        tags:
          - "{{ .Tag }}"
```

```bash
# Ensure the image exists locally
docker images myapp:latest

# Push it
lazyoci build --tag v1.0.0
```

## Use Tag Templates

Tag values support Go template variables resolved at build time.

### Available variables

```yaml
tags:
  - "{{ .Tag }}"               # --tag flag or LAZYOCI_TAG env var
  - "{{ .GitSHA }}"            # git short SHA
  - "{{ .GitBranch }}"         # git branch name
  - "{{ .ChartVersion }}"      # Chart.yaml version (helm only)
  - "{{ .Timestamp }}"         # YYYYMMDDHHmmss UTC
  - "{{ .Version }}"           # semver from git tag (1.2.3)
  - "{{ .VersionMajorMinor }}" # major.minor (1.2)
  - latest                     # literal string (no template)
```

### Combine variables

```yaml
tags:
  - "{{ .Tag }}-{{ .GitSHA }}"
  - "{{ .GitBranch }}-{{ .Timestamp }}"
```

## Use Semver from Git Tags

The `{{ .Version }}` family auto-detects the version from your git tags. No need to hard-code version numbers or pass `--tag` for semver tagging.

### Automatic version detection

If your repo has a git tag like `v1.2.3`:

```yaml
tags:
  - "{{ .Version }}"           # -> "1.2.3"
  - "{{ .VersionMajorMinor }}" # -> "1.2"
  - "{{ .VersionMajor }}"      # -> "1"
```

```bash
# Just run build — version is detected from git tag
lazyoci build
```

### Explicit version override

Override with `--tag` (if it's a valid semver) or the `LAZYOCI_VERSION` env var:

```bash
# --tag populates both {{ .Tag }} and {{ .Version }}
lazyoci build --tag v2.0.0

# Env var overrides git detection
LAZYOCI_VERSION=3.0.0-rc.1 lazyoci build
```

### Prerelease tags

For prerelease versions like `v2.0.0-rc.1`:

```yaml
tags:
  - "{{ .Version }}-{{ .VersionPrerelease }}"  # -> "2.0.0-rc.1"
  - "{{ .Version }}"                            # -> "2.0.0"
```

## Use in CI/CD

lazyoci supports environment variables for CI systems where flags are inconvenient.

### Environment variables

| Variable | Fallback for | Description |
|----------|--------------|-------------|
| `LAZYOCI_TAG` | `--tag` | Sets `{{ .Tag }}` when flag not specified |
| `LAZYOCI_VERSION` | Git tag detection | Overrides `{{ .Version }}` family |

### GitHub Actions example

```yaml
# .github/workflows/release.yml
on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: mistergrinvalds/lazyoci@v1
        with:
          tag: ${{ github.ref_name }}
```

The `{{ .Tag }}` and `{{ .Version }}` variables are automatically populated from the git tag that triggered the workflow.

### Any CI system

```bash
# Set via env vars (works in any CI)
export LAZYOCI_TAG="${CI_TAG}"
export LAZYOCI_VERSION="${CI_TAG}"
lazyoci build
```

## Preview with Dry Run

See what would be built and pushed without actually doing it.

### Dry run

```bash
lazyoci build --tag v1.0.0 --dry-run
```

### Dry run with JSON output

```bash
lazyoci build --tag v1.0.0 --dry-run -o json
```

## Build Without Pushing

Build artifacts locally without pushing to registries.

```bash
lazyoci build --tag v1.0.0 --no-push
```

## Build a Specific Artifact

When a `.lazy` file contains multiple artifacts, build only one.

### Filter by name

```bash
lazyoci build --tag v1.0.0 --artifact myapp
```

### Filter by type

```bash
# Build all helm artifacts
lazyoci build --tag v1.0.0 --artifact helm
```

### Filter by index

```bash
# Build the first artifact (0-based)
lazyoci build --tag v1.0.0 --artifact 0
```

## Push to Multiple Registries

Push the same artifact to multiple registries by adding multiple targets.

```yaml
targets:
  - registry: ghcr.io/owner/myapp
    tags:
      - "{{ .Tag }}"
  - registry: docker.io/owner/myapp
    tags:
      - "{{ .Tag }}"
      - latest
```

## Push to Insecure Registries

Push to HTTP registries (e.g., local development).

```bash
lazyoci build --tag v1.0.0 --insecure
```

Or use the local development registry:

```bash
make registry-up
lazyoci build examples/artifact --tag v1.0.0 --insecure
```

:::tip Examples
Complete working examples for every artifact type are available in the [`examples/`](https://github.com/mistergrinvalds/lazyoci/tree/main/examples) directory.
:::
