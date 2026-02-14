---
title: CI/CD with GitHub Actions
description: Set up lazyoci in GitHub Actions to build and push OCI artifacts on release
sidebar_position: 6
---

# CI/CD with GitHub Actions

Set up lazyoci in your GitHub Actions workflows to automatically build and push OCI artifacts when you create releases or push tags.

## Use the lazyoci GitHub Action

The simplest way to use lazyoci in CI is with the official composite action.

### Basic release workflow

```yaml
# .github/workflows/release.yml
name: Build and Push

on:
  push:
    tags: ['v*']

permissions:
  contents: read
  packages: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for git describe

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: mistergrinvalds/lazyoci@v1
        id: build
        with:
          tag: ${{ github.ref_name }}

      - name: Print pushed references
        run: echo "${{ steps.build.outputs.references }}"
```

When you push a tag like `v1.2.3`, this workflow:

1. Checks out the repo with full git history
2. Authenticates to GitHub Container Registry
3. Runs `lazyoci build` with `--tag v1.2.3`
4. All template variables are populated automatically:
   - `{{ .Tag }}` = `v1.2.3`
   - `{{ .Version }}` = `1.2.3`
   - `{{ .VersionMajor }}` = `1`
   - `{{ .GitSHA }}` = commit SHA

### Action inputs

| Input | Default | Description |
|-------|---------|-------------|
| `tag` | Auto from `GITHUB_REF_NAME` | Tag value for `{{ .Tag }}` |
| `version` | `""` | Explicit version override for `{{ .Version }}` |
| `push` | `true` | Push to registries after building |
| `artifact` | `""` | Filter by artifact name, type, or index |
| `platform` | `""` | Comma-separated platforms (e.g., `linux/amd64,linux/arm64`) |
| `config` | `""` | Path to `.lazy` config file or directory |
| `insecure` | `false` | Allow HTTP registries |
| `dry-run` | `false` | Preview without building/pushing |
| `lazyoci-version` | `latest` | lazyoci version to install |

### Action outputs

| Output | Description |
|--------|-------------|
| `references` | Newline-separated pushed references |
| `digests` | Newline-separated manifest digests |
| `result` | Full JSON build result |

## Use Environment Variables

For CI systems where the action isn't available, use environment variables.

### Environment variable reference

| Variable | Fallback for | Description |
|----------|--------------|-------------|
| `LAZYOCI_TAG` | `--tag` flag | Sets `{{ .Tag }}` when not specified on CLI |
| `LAZYOCI_VERSION` | Git tag detection | Overrides `{{ .Version }}` and all semver fields |

### Generic CI script

```bash
#!/bin/bash
set -euo pipefail

# Install lazyoci
go install github.com/mistergrinvalds/lazyoci/cmd/lazyoci@latest

# Set tag from CI environment
export LAZYOCI_TAG="${CI_COMMIT_TAG:-}"
export LAZYOCI_VERSION="${CI_COMMIT_TAG:-}"

# Build and push
lazyoci build
```

## Registry Authentication in CI

### GitHub Container Registry (GHCR)

```yaml
- uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

### Docker Hub

```yaml
- uses: docker/login-action@v3
  with:
    username: ${{ secrets.DOCKERHUB_USERNAME }}
    password: ${{ secrets.DOCKERHUB_TOKEN }}
```

### AWS ECR

```yaml
- uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: arn:aws:iam::role/ecr-push
    aws-region: us-east-1
- uses: aws-actions/amazon-ecr-login@v2
```

### DigitalOcean Container Registry

```yaml
- uses: digitalocean/action-doctl@v2
  with:
    token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
- run: doctl registry login
```

## Common Workflow Patterns

### Build on tag push, skip on PR

```yaml
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

### Dry run on PR, push on tag

```yaml
on:
  pull_request:
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
        if: github.ref_type == 'tag'
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: mistergrinvalds/lazyoci@v1
        with:
          tag: ${{ github.ref_name || 'pr-check' }}
          push: ${{ github.ref_type == 'tag' }}
          dry-run: ${{ github.ref_type != 'tag' }}
```

### Manual dispatch with version input

```yaml
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to build (e.g., v1.2.3)'
        required: true
        type: string

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
          tag: ${{ inputs.version }}
```

### Build specific artifact only

```yaml
- uses: mistergrinvalds/lazyoci@v1
  with:
    tag: ${{ github.ref_name }}
    artifact: api-server  # Only build the "api-server" artifact
```

### Multi-platform image build

```yaml
- uses: docker/setup-qemu-action@v3  # Required for cross-platform
- uses: mistergrinvalds/lazyoci@v1
  with:
    tag: ${{ github.ref_name }}
    platform: "linux/amd64,linux/arm64"
```

## Recommended .lazy Config for CI

A `.lazy` file designed for CI with semver tagging:

```yaml
version: 1
artifacts:
  - type: image
    name: myapp
    dockerfile: Dockerfile
    context: "."
    platforms:
      - linux/amd64
      - linux/arm64
    targets:
      - registry: ghcr.io/owner/myapp
        tags:
          - "{{ .Version }}"           # 1.2.3
          - "{{ .VersionMajorMinor }}" # 1.2
          - "{{ .GitSHA }}"            # abc1234
          - latest
```

This config produces 4 tags per release without any hard-coded versions. The version is automatically detected from the git tag that triggered the workflow.

:::tip fetch-depth: 0
Always use `fetch-depth: 0` with `actions/checkout` so that `git describe --tags` can find your tags. Without full history, version auto-detection won't work.
:::
