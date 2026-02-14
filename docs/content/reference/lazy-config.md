---
title: .lazy Config File
sidebar_position: 4
---

# .lazy Config File

The `.lazy` file is a YAML configuration that defines OCI artifacts to build and push. It sits alongside the build context (Dockerfiles, chart directories, etc.) and is processed by `lazyoci build`.

## Schema

```yaml
version: 1                    # Required. Must be 1.
artifacts:                     # Required. List of artifacts to build.
  - type: <string>             # Required. One of: image, helm, artifact, docker.
    name: <string>             # Optional. Human-readable name for output/filtering.
    targets:                   # Required. At least one push target.
      - registry: <string>    # Required. Registry/repository path.
        tags:                  # Required. At least one tag.
          - <string>           # Tag value. Supports template variables.

    # type: image
    dockerfile: <string>       # Default: "Dockerfile"
    context: <string>          # Default: "."
    platforms:                 # Optional. Multi-arch build targets.
      - <string>               # e.g., "linux/amd64"
    buildArgs:                 # Optional. Docker build arguments.
      KEY: value

    # type: helm
    chartPath: <string>        # Required. Path to chart directory.

    # type: artifact
    mediaType: <string>        # Optional. OCI artifactType annotation.
    files:                     # Required. Files to include.
      - path: <string>         # Required. File path relative to .lazy.
        mediaType: <string>    # Required. OCI media type for this file.

    # type: docker
    image: <string>            # Required. Docker daemon image reference.
```

## Artifact Types

### `image` -- Container Image

Builds a container image from a Dockerfile using `docker buildx build`. The output is an OCI layout tarball (`--output type=oci`), which is then pushed to each target.

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `dockerfile` | No | `Dockerfile` | Path to Dockerfile |
| `context` | No | `.` | Build context directory |
| `platforms` | No | Host platform | Target platforms for multi-arch |
| `buildArgs` | No | `{}` | `--build-arg` key-value pairs |

**Prerequisites:** `docker buildx` must be available.

```yaml
- type: image
  name: myapp
  dockerfile: Dockerfile
  context: "."
  platforms:
    - linux/amd64
    - linux/arm64
  buildArgs:
    GO_VERSION: "1.23"
  targets:
    - registry: ghcr.io/owner/myapp
      tags:
        - "{{ .Tag }}"
        - latest
```

### `helm` -- Helm Chart

Packages a Helm chart directory as an OCI artifact with the standard Helm media types. No Helm CLI is required -- lazyoci creates the tar.gz and OCI manifest directly.

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `chartPath` | Yes | -- | Path to chart directory containing `Chart.yaml` |

The chart metadata (name, version, description) is read from `Chart.yaml` and included in the OCI config blob.

**OCI media types used:**
- Config: `application/vnd.cncf.helm.config.v1+json`
- Layer: `application/vnd.cncf.helm.chart.content.v1.tar+gzip`

```yaml
- type: helm
  name: mychart
  chartPath: charts/mychart
  targets:
    - registry: ghcr.io/owner/charts/mychart
      tags:
        - "{{ .ChartVersion }}"
        - latest
```

### `artifact` -- Generic OCI Artifact

Packages arbitrary files as an OCI artifact with custom media types. Useful for configs, policies, WASM modules, or any non-image content.

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `mediaType` | No | `application/vnd.unknown.artifact.v1` | `artifactType` annotation |
| `files` | Yes | -- | Files to include as layers |
| `files[].path` | Yes | -- | File path relative to `.lazy` |
| `files[].mediaType` | Yes | -- | OCI media type for this layer |

```yaml
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

### `docker` -- Docker Daemon Image

Exports an existing image from the local Docker daemon and pushes it to a registry. No build step is needed -- the image must already exist in Docker.

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `image` | Yes | -- | Docker daemon image reference (e.g., `myapp:latest`) |

The image is exported via `docker save`, converted from Docker save format to OCI layout, then pushed.

```yaml
- type: docker
  name: local-nginx
  image: nginx:alpine
  targets:
    - registry: ghcr.io/owner/nginx-mirror
      tags:
        - "{{ .Tag }}"
        - latest
```

## Template Variables

Tag values support Go template syntax. Variables are resolved at build time.

### General variables

| Variable | Source | Example |
|----------|--------|---------|
| `{{ .Tag }}` | `--tag` CLI flag or `LAZYOCI_TAG` env var | `v1.0.0` |
| `{{ .GitSHA }}` | `git rev-parse --short HEAD` | `e5bce6f` |
| `{{ .GitBranch }}` | `git branch --show-current` | `main` |
| `{{ .ChartVersion }}` | `Chart.yaml` version field | `0.1.0` |
| `{{ .Timestamp }}` | UTC build time (`YYYYMMDDHHmmss`) | `20260209153000` |

### Semver variables

These variables are auto-populated from git tags (or overridden via `LAZYOCI_VERSION` env var or a semver `--tag` value).

| Variable | Description | Example |
|----------|-------------|---------|
| `{{ .Version }}` | Clean semver (`v` prefix stripped) | `1.2.3` |
| `{{ .VersionMajor }}` | Major version component | `1` |
| `{{ .VersionMinor }}` | Minor version component | `2` |
| `{{ .VersionPatch }}` | Patch version component | `3` |
| `{{ .VersionPrerelease }}` | Prerelease identifier (empty if none) | `rc.1` |
| `{{ .VersionMajorMinor }}` | `MAJOR.MINOR` shorthand | `1.2` |
| `{{ .VersionRaw }}` | Raw version string before parsing | `v1.2.3-rc.1` |

**Version resolution priority:**

1. `LAZYOCI_VERSION` environment variable (highest priority)
2. `--tag` flag value (if it's a valid semver)
3. `git describe --tags --abbrev=0` (nearest git tag)

### Notes

- `{{ .ChartVersion }}` is only available for `helm` type artifacts. Using it on other types will produce an empty string.
- Tags without template delimiters (`{{ }}`) are used as literal values.
- If a version source is not valid semver, only `{{ .Version }}` and `{{ .VersionRaw }}` are populated; component fields remain empty.

### Common tag patterns

```yaml
tags:
  # Semver: push as 1.2.3, 1.2, and latest
  - "{{ .Version }}"
  - "{{ .VersionMajorMinor }}"
  - latest

  # Git-based: commit SHA for traceability
  - "{{ .GitSHA }}"

  # Combined: version + SHA for immutable + readable
  - "{{ .Version }}-{{ .GitSHA }}"

  # Prerelease-aware
  - "{{ .Version }}{{ if .VersionPrerelease }}-{{ .VersionPrerelease }}{{ end }}"
```

## Targets

Each artifact must have at least one target. A target specifies a registry and one or more tags.

```yaml
targets:
  - registry: ghcr.io/owner/myapp
    tags:
      - "{{ .Tag }}"
      - "{{ .GitSHA }}"
      - latest
  - registry: docker.io/owner/myapp
    tags:
      - "{{ .Tag }}"
```

The `registry` field is the full repository path excluding the tag. Each tag is pushed as a separate reference.

## Path Resolution

All relative paths in the `.lazy` file (dockerfile, context, chartPath, files) are resolved relative to the directory containing the `.lazy` file.

```
project/
├── .lazy              # config references paths relative to this location
├── Dockerfile         # dockerfile: Dockerfile
├── charts/
│   └── myapp/         # chartPath: charts/myapp
└── config/
    └── app.json       # files: [{path: config/app.json}]
```

## Multiple Artifacts

A single `.lazy` file can define multiple artifacts. They are built and pushed in order.

```yaml
version: 1
artifacts:
  - type: image
    name: api-server
    # ...
  - type: helm
    name: api-chart
    # ...
  - type: artifact
    name: api-spec
    # ...
```

Use `--artifact` to build a specific one:

```bash
lazyoci build --tag v1.0.0 --artifact api-chart
```

## Validation Rules

The config is validated before any build starts:

- `version` must be `1`
- At least one artifact is required
- Each artifact must have a valid `type`
- Each artifact must have at least one target with a registry and tags
- `helm` artifacts require `chartPath`
- `artifact` type requires `files` with `path` and `mediaType` on each entry
- `docker` type requires `image`

## Examples

Complete working examples are available in the [`examples/`](https://github.com/mistergrinvalds/lazyoci/tree/main/examples) directory:

- [`examples/image/`](https://github.com/mistergrinvalds/lazyoci/tree/main/examples/image) -- container image with multi-arch
- [`examples/helm/`](https://github.com/mistergrinvalds/lazyoci/tree/main/examples/helm) -- Helm chart packaging
- [`examples/artifact/`](https://github.com/mistergrinvalds/lazyoci/tree/main/examples/artifact) -- generic OCI artifact
- [`examples/docker/`](https://github.com/mistergrinvalds/lazyoci/tree/main/examples/docker) -- Docker daemon image push
- [`examples/multi/`](https://github.com/mistergrinvalds/lazyoci/tree/main/examples/multi) -- multiple artifacts in one config
