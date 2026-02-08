---
title: Pull Multi-platform Images
description: Target specific architectures and operating systems when pulling artifacts
sidebar_position: 6
---

# Pull Multi-platform Images

Pull container artifacts for specific platforms using the `--platform` flag. Useful for cross-platform development and deployment.

## Understand Platform Format

Platforms are specified as `os/architecture[/variant]`.

### Common platforms

- `linux/amd64` - Intel/AMD 64-bit Linux
- `linux/arm64` - ARM 64-bit Linux (Apple Silicon, AWS Graviton)
- `linux/arm/v7` - ARM 32-bit Linux (Raspberry Pi)
- `windows/amd64` - Windows 64-bit
- `darwin/amd64` - Intel Mac
- `darwin/arm64` - Apple Silicon Mac

## Pull for Specific Platform

Override the default platform detection.

### Pull ARM64 image

```bash
lazyoci pull --platform linux/arm64 alpine:latest
```

This pulls the ARM64 variant regardless of your current system architecture.

### Pull for different OS

```bash
lazyoci pull --platform windows/amd64 mcr.microsoft.com/windows/servercore:ltsc2022
```

### Pull with variant

```bash
lazyoci pull --platform linux/arm/v7 alpine:latest
```

For specific ARM variants like ARMv7.

## Cross-Platform Development

Pull images for target deployment platforms.

### Develop on Intel, deploy on ARM

```bash
# Pull ARM image for testing on Intel Mac
lazyoci pull --platform linux/arm64 nginx:latest --dest ~/arm64-nginx

# Inspect the pulled artifact
ls -la ~/arm64-nginx/
```

### Test multiple architectures

```bash
# Pull same image for different platforms
lazyoci pull --platform linux/amd64 postgres:15 --dest ~/postgres-amd64
lazyoci pull --platform linux/arm64 postgres:15 --dest ~/postgres-arm64

# Compare manifest differences
cat ~/postgres-amd64/index.json
cat ~/postgres-arm64/index.json
```

## Docker Integration

Pull platform-specific images directly into Docker.

### Load ARM64 image into Docker

```bash
lazyoci pull --platform linux/arm64 redis:latest --docker
```

Docker will store this as the ARM64 variant, even on x86 systems.

### Verify platform in Docker

```bash
docker inspect redis:latest | grep Architecture
```

Should show "arm64" even if you're running on an Intel system.

## Platform Auto-Detection

When no platform is specified, lazyoci detects your current platform.

### Check current platform

```bash
# Pull without platform (uses auto-detection)
lazyoci pull alpine:latest --quiet

# Manual check of what was pulled
uname -m  # Shows your architecture
```

### Override auto-detection

```bash
# Force specific platform
lazyoci pull --platform linux/amd64 alpine:latest
```

Useful when you need a specific platform regardless of your current system.

## Multi-Arch Images

Some images support multiple architectures in a single tag.

### Check available platforms

```bash
# Use skopeo to inspect available platforms
skopeo inspect docker://alpine:latest | jq '.RepoTags'
```

Or use Docker:

```bash
docker manifest inspect alpine:latest
```

### Pull specific variant from multi-arch

```bash
# Pull specific architecture from multi-arch image
lazyoci pull --platform linux/arm64 alpine:latest
lazyoci pull --platform linux/amd64 alpine:latest
```

Both will have the same tag but different underlying images.

## Cloud Deployment Scenarios

### AWS Graviton instances

```bash
# Pull ARM64 images for Graviton processors
lazyoci pull --platform linux/arm64 node:18 --docker
lazyoci pull --platform linux/arm64 nginx:latest --docker
```

### Raspberry Pi deployment

```bash
# Pull ARMv7 images for Raspberry Pi
lazyoci pull --platform linux/arm/v7 alpine:latest --dest ~/pi-images/
```

### Apple Silicon development

```bash
# Pull ARM64 images for M1/M2 Macs
lazyoci pull --platform linux/arm64 mysql:8.0 --docker
```

## Troubleshoot Platform Issues

### Platform not available

If you get a "platform not supported" error:

```bash
# Check what platforms are available
skopeo inspect docker://your-image:tag
```

Look for the "Digest" field in different manifests.

### Wrong architecture pulled

Verify the pulled image architecture:

```bash
# Pull to directory and inspect
lazyoci pull --platform linux/arm64 nginx:latest --dest /tmp/nginx-check

# Check the manifest
cat /tmp/nginx-check/blobs/sha256/$(cat /tmp/nginx-check/index.json | jq -r '.manifests[0].digest' | cut -d: -f2) | jq '.architecture'
```

### Docker platform mismatch

When loading into Docker on a different architecture:

```bash
# Enable Docker experimental features for cross-platform
export DOCKER_CLI_EXPERIMENTAL=enabled

# Pull and load cross-platform image
lazyoci pull --platform linux/arm64 alpine:latest --docker
```

## Batch Platform Operations

### Pull for multiple platforms

```bash
#!/bin/bash
platforms=("linux/amd64" "linux/arm64" "linux/arm/v7")
image="alpine:latest"

for platform in "${platforms[@]}"; do
  echo "Pulling $image for $platform..."
  lazyoci pull --platform "$platform" "$image" --dest "/tmp/${platform//\//-}-alpine" --quiet
done
```

### Compare platform variants

```bash
# Pull same image for different platforms
lazyoci pull --platform linux/amd64 nginx:latest --dest /tmp/nginx-amd64
lazyoci pull --platform linux/arm64 nginx:latest --dest /tmp/nginx-arm64

# Compare sizes
du -sh /tmp/nginx-*
```

## Platform-Specific Storage

### Organize by platform

```bash
# Use platform in destination path
platform="linux/arm64"
safe_platform="${platform//\//-}"  # Replace / with -

lazyoci pull --platform "$platform" redis:latest --dest "~/artifacts/$safe_platform/redis"
```

### Environment-specific caching

```bash
# Set platform-specific cache directory
export LAZYOCI_ARTIFACT_DIR="~/.cache/lazyoci-arm64"
lazyoci pull --platform linux/arm64 postgres:15 --docker
```

## Verify Platform Configuration

### Test platform detection

```bash
# Check what platform lazyoci detects
lazyoci pull alpine:latest --dest /tmp/auto-platform
cat /tmp/auto-platform/index.json | jq '.manifests[0].platform'
```

### Confirm cross-platform pull

```bash
# Pull opposite architecture and verify
current_arch=$(uname -m)
if [ "$current_arch" = "x86_64" ]; then
  target_platform="linux/arm64"
else
  target_platform="linux/amd64"
fi

lazyoci pull --platform "$target_platform" hello-world:latest --dest /tmp/cross-platform
cat /tmp/cross-platform/index.json | jq '.manifests[0].platform'
```

The platform in the manifest should match your `--platform` flag, not your system architecture.

:::tip Platform Strategy
For production deployments, always explicitly specify the target platform with `--platform` rather than relying on auto-detection. This ensures consistent behavior across different development environments.
:::