---
title: Push Artifacts to Registries
description: Push built artifacts to OCI-compliant container registries
sidebar_position: 5
---

# Push Artifacts to Registries

Push OCI artifacts to container registries using `lazyoci build`. This guide covers authentication, multi-registry setups, and common registry configurations.

## Authentication for Push

lazyoci uses the same credential chain for push as for pull. If you can pull from a registry, you can push to it.

### Use existing Docker credentials

```bash
# Log in to your registry
docker login ghcr.io

# lazyoci automatically uses these credentials
lazyoci build --tag v1.0.0
```

### Use lazyoci credentials

```bash
# Add credentials via lazyoci
lazyoci registry add ghcr.io --user username --pass token

# Push
lazyoci build --tag v1.0.0
```

## Push to GitHub Container Registry

### Create a personal access token

1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Create a token with `write:packages` scope
3. Log in:

```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

### .lazy config for GHCR

```yaml
version: 1
artifacts:
  - type: image
    name: myapp
    targets:
      - registry: ghcr.io/owner/myapp
        tags:
          - "{{ .Tag }}"
          - "{{ .GitSHA }}"
```

```bash
lazyoci build --tag v1.0.0
```

## Push to Docker Hub

### Log in to Docker Hub

```bash
docker login
```

### .lazy config for Docker Hub

```yaml
targets:
  - registry: docker.io/username/myapp
    tags:
      - "{{ .Tag }}"
      - latest
```

## Push to AWS ECR

### Authenticate with ECR

```bash
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com
```

### .lazy config for ECR

```yaml
targets:
  - registry: 123456789.dkr.ecr.us-east-1.amazonaws.com/myapp
    tags:
      - "{{ .Tag }}"
```

## Push to a Local Registry

### Start the development registry

```bash
make registry-up  # starts localhost:5050
```

### Push with insecure flag

```bash
lazyoci build --tag v1.0.0 --insecure
```

### Verify the push

```bash
lazyoci browse tags localhost:5050/examples/myapp
```

## Push to Multiple Registries

A single `.lazy` file can push to multiple registries with different tags per registry.

```yaml
targets:
  - registry: ghcr.io/owner/myapp
    tags:
      - "{{ .Tag }}"
      - "{{ .GitSHA }}"
  - registry: docker.io/owner/myapp
    tags:
      - "{{ .Tag }}"
      - latest
  - registry: 123456789.dkr.ecr.us-east-1.amazonaws.com/myapp
    tags:
      - "{{ .Tag }}"
```

All targets are pushed sequentially after the build completes.

## Verify a Push

After pushing, verify the artifact is accessible.

### Browse tags

```bash
lazyoci browse tags ghcr.io/owner/myapp
```

### Inspect manifest

```bash
lazyoci browse manifest ghcr.io/owner/myapp:v1.0.0
```

### Pull back

```bash
lazyoci pull ghcr.io/owner/myapp:v1.0.0
```

## Troubleshoot Push Failures

### Authentication errors

```bash
# Test registry connectivity
lazyoci registry test ghcr.io

# Verify credentials work for pull
lazyoci browse tags ghcr.io/owner/myapp
```

### Permission denied

Ensure your token has write/push scope:
- **GHCR**: `write:packages` permission
- **Docker Hub**: Read/Write access token
- **ECR**: `ecr:PutImage`, `ecr:InitiateLayerUpload`, `ecr:UploadLayerPart`, `ecr:CompleteLayerUpload` permissions

### Insecure registry errors

If pushing to an HTTP registry, pass `--insecure`:

```bash
lazyoci build --tag v1.0.0 --insecure
```

:::tip Dry Run First
Always preview with `--dry-run` before pushing to production registries to verify tag templates resolve correctly:
```bash
lazyoci build --tag v1.0.0 --dry-run
```
:::
