---
title: Build System Architecture
sidebar_position: 7
---

# Build System Architecture

The `lazyoci build` command provides a declarative way to build and push OCI artifacts. Understanding its architecture explains the design decisions behind the `.lazy` config format, the choice of build tools, and how artifacts flow from source to registry.

## Design Philosophy

### Declarative over imperative

The `.lazy` config file describes **what** to build, not **how** to build it. This mirrors the approach of tools like Docker Compose and Kubernetes manifests. A developer declares their artifacts, targets, and tags, and lazyoci handles the build orchestration.

This stands in contrast to shell scripts that chain together `docker build`, `docker tag`, and `docker push` commands. The declarative approach is easier to review, harder to get wrong, and produces consistent results.

### Leverage existing tools

Rather than implementing a container image builder from scratch, lazyoci shells out to `docker buildx build` for image builds. This provides several advantages:

**BuildKit integration**: `docker buildx` uses BuildKit under the hood, providing layer caching, multi-stage builds, and multi-architecture support. Reimplementing these features would be impractical.

**User familiarity**: Developers already understand Dockerfiles, build contexts, and build arguments. lazyoci doesn't introduce new build syntax.

**Zero new dependencies**: By shelling out to `docker buildx`, lazyoci avoids pulling in the BuildKit Go SDK (~230 transitive dependencies) or the Docker Engine SDK.

### Native packaging where practical

For Helm charts and generic artifacts, lazyoci packages them directly in Go rather than shelling out to external tools:

**Helm charts**: Manual tar.gz packaging + OCI manifest construction via oras-go. This avoids requiring the Helm CLI and gives full control over the OCI media types.

**Generic artifacts**: Direct `oras.PackManifest()` calls with user-specified media types. No external tool exists for this use case.

**Docker daemon images**: `docker save` + format conversion. The conversion from Docker save format to OCI layout reuses the same code path as the pull system's `LoadToDocker` (but in reverse).

## Architecture

### Build pipeline

Each artifact goes through a three-stage pipeline:

```
Config Parse → Type-specific Build → Push to Targets
```

**Stage 1: Config Parse** reads the `.lazy` file, validates the schema, resolves template variables (git SHA, branch, timestamp), and renders tags.

**Stage 2: Build** dispatches to a type-specific handler that produces an OCI layout on disk:

| Type | Build method | Output |
|------|-------------|--------|
| `image` | `docker buildx build --output type=oci` | OCI layout tarball (extracted) |
| `helm` | `tar.gz` + `oras.PackManifest()` | OCI layout via oras memory store |
| `artifact` | `oras.PackManifest()` | OCI layout via oras memory store |
| `docker` | `docker save` + format conversion | OCI layout from converted Docker save |

**Stage 3: Push** copies the OCI layout to each target registry using `oras.Copy()`. The same push logic handles all four artifact types uniformly.

### OCI layout as intermediate format

Every build handler produces an OCI layout directory as its output. This is the key architectural decision that enables a uniform push path:

```
image handler  ──┐
helm handler   ──┤── OCI layout on disk ──→ oras.Copy() ──→ registry
artifact handler─┤
docker handler ──┘
```

The OCI layout is a standard on-disk format defined by the OCI Image Spec. It contains `index.json`, `oci-layout`, and content-addressed blobs. Any tool that understands OCI layout can consume this intermediate output.

### Shared utilities

The build system shares infrastructure with the pull system through the `pkg/ociutil` package:

- **`ociutil.ParseReference`**: Reference parsing (`registry/repo:tag`)
- **`ociutil.NewRemoteRepository`**: oras remote repository creation with auth
- **`ociutil.OCIIndex`, `OCIManifest`**: OCI layout types
- **`ociutil.DockerSaveManifest`**: Docker save format types

This sharing was achieved by extracting these utilities from `pkg/pull` during the build system implementation, ensuring both pull and push use identical reference parsing and registry connection logic.

## Why shell out to docker buildx?

Three alternatives were considered for building container images:

### Docker Engine SDK (`github.com/docker/docker/client`)

The `ImageBuild()` API connects to the Docker daemon and returns a build stream. However, it uses Docker's **legacy builder** by default, not BuildKit. Getting BuildKit behavior through this API is awkward, and multi-architecture builds require building each platform separately and assembling manifest lists manually.

### BuildKit Go SDK (`github.com/moby/buildkit`)

Direct BuildKit integration would provide the most control, but it requires a running `buildkitd` daemon and adds ~230 transitive dependencies (~50-100MB to the binary). This is disproportionate for a TUI tool.

### Shell out to `docker buildx build`

This is the chosen approach. `docker buildx build` with `--output type=oci,dest=output.tar` provides:

- BuildKit under the hood (Docker 23+)
- Multi-platform via `--platform` flag
- Direct OCI layout output (no conversion needed)
- Zero new Go dependencies
- Users already have Docker installed

The trade-off is a runtime dependency on the `docker` CLI, which is acceptable since lazyoci already depends on Docker for its `--docker` pull flag.

## Why oras-go for push?

Two libraries were considered for pushing to registries:

**oras-go** (already a dependency) provides `oras.Copy()` which can copy from a local OCI store to a remote repository. It handles authentication, chunked uploads, and manifest pushing.

**go-containerregistry** (ggcr) provides excellent multi-architecture support and Docker daemon integration, but would add a new dependency.

The decision was to **use oras-go exclusively**, since `docker buildx` already handles multi-arch image assembly. The push step just needs to copy an OCI layout to a registry, which oras-go does well.

## Template variable resolution

Tag templates use Go's `text/template` package with `missingkey=error` to catch undefined variables early. Variables are resolved once per artifact before the build starts:

```
.Tag          ← --tag CLI flag
.GitSHA       ← git rev-parse --short HEAD
.GitBranch    ← git branch --show-current
.ChartVersion ← Chart.yaml version field (helm only)
.Timestamp    ← time.Now().UTC().Format("20060102150405")
```

Git values are resolved by executing `git` commands. If git is not available or the directory isn't a repository, these variables resolve to empty strings rather than causing an error.

## Docker save to OCI layout conversion

The `type: docker` handler needs to convert Docker save format to OCI layout. This is the inverse of the conversion that `lazyoci pull --docker` performs:

**Pull direction**: OCI layout → Docker save format → `docker load`
**Push direction**: `docker save` → Docker save format → OCI layout → `oras.Copy()`

The conversion involves:
1. Extracting the Docker save tarball
2. Reading `manifest.json` to find config and layer paths
3. Gzip-compressing uncompressed layers (Docker save stores them uncompressed)
4. Computing SHA-256 digests for all blobs
5. Constructing an OCI manifest and index.json

The shared types (`DockerSaveManifest`, `OCIIndex`, `OCIManifest`) in `pkg/ociutil/convert.go` are used by both directions of conversion.
