---
title: Pull Images to Docker
description: Download artifacts and load them directly into Docker
sidebar_position: 5
---

# Pull Images to Docker

Pull container artifacts and automatically load them into your local Docker daemon for immediate use.

## Basic Docker Pull

Pull an artifact and load it into Docker in one command.

### Pull and load image

```bash
lazyoci pull nginx:latest --docker
```

The `--docker` flag downloads the artifact and loads it into Docker automatically.

### Verify image is available

```bash
docker images nginx
```

You should see the nginx image listed with the "latest" tag.

### Run the pulled image

```bash
docker run --rm nginx:latest nginx -v
```

## Pull with Specific Destination

Control where the artifact is downloaded before loading into Docker.

### Pull to specific directory

```bash
lazyoci pull alpine:latest --docker --dest ~/my-images/alpine
```

This downloads the artifact to `~/my-images/alpine` then loads it into Docker.

### Keep artifact after loading

The artifact remains in the destination directory for inspection or reuse:

```bash
# Explore the OCI layout
ls -la ~/my-images/alpine/
```

## Quiet Mode

Suppress output during the pull operation.

### Silent pull

```bash
lazyoci pull busybox:latest --docker --quiet
```

Only errors will be displayed. Useful for scripting or automated workflows.

## Platform-Specific Pulls

Pull artifacts for specific platforms when loading into Docker.

### Pull for current platform

```bash
# Automatically selects appropriate platform
lazyoci pull --platform linux/amd64 node:18 --docker
```

### Pull multi-arch image

```bash
# Let Docker choose the right variant
lazyoci pull --platform linux/arm64 alpine:latest --docker
```

See [Multi-platform Images](multi-platform-images.md) for more platform options.

## Private Registry Images

Pull from authenticated registries into Docker.

### Pull private image

```bash
# Ensure you're logged in first
docker login your-registry.com

# Pull private image
lazyoci pull your-registry.com/private/app:v1.0 --docker
```

### Pull with explicit credentials

```bash
# Add registry credentials
lazyoci registry add your-registry.com --user username --pass password

# Pull and load
lazyoci pull your-registry.com/private/app:v1.0 --docker
```

## Optimization with Skopeo

lazyoci automatically uses skopeo for faster Docker loading when available.

### Install skopeo (recommended)

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install skopeo
```

**CentOS/RHEL:**
```bash
sudo dnf install skopeo
```

**macOS (Homebrew):**
```bash
brew install skopeo
```

### Verify optimization is working

```bash
# Pull with verbose output to see skopeo usage
lazyoci pull alpine:latest --docker
```

You should see a message indicating skopeo is being used for the Docker load operation.

## Batch Operations

Pull multiple images into Docker efficiently.

### Pull multiple images

```bash
# Pull several images
lazyoci pull nginx:latest --docker --quiet
lazyoci pull alpine:latest --docker --quiet  
lazyoci pull postgres:13 --docker --quiet

# Verify all images are loaded
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}"
```

### Script for bulk pulls

```bash
#!/bin/bash
images=(
  "nginx:latest"
  "alpine:3.18"
  "redis:7"
  "postgres:15"
)

for image in "${images[@]}"; do
  echo "Pulling $image..."
  lazyoci pull "$image" --docker --quiet
done

echo "All images pulled successfully"
docker images
```

## Troubleshoot Loading Issues

### Check Docker daemon

```bash
# Verify Docker is running
docker info
```

### Test without Docker flag

```bash
# Pull without loading to isolate issues
lazyoci pull problematic-image:tag --dest /tmp/test-pull
```

If this succeeds but `--docker` fails, the issue is with Docker loading, not the pull operation.

### Manual Docker load

```bash
# Pull to directory first
lazyoci pull nginx:latest --dest /tmp/nginx-oci

# Attempt manual load with skopeo
skopeo copy oci:/tmp/nginx-oci docker-daemon:nginx:latest
```

### Check skopeo installation

```bash
# Verify skopeo is available
skopeo --version

# Test skopeo Docker access
skopeo inspect docker-daemon:hello-world:latest
```

## Performance Tips

### Use local caching

```bash
# Set custom artifact directory for better performance
export LAZYOCI_ARTIFACT_DIR=/fast-ssd/lazyoci-cache
lazyoci pull large-image:latest --docker
```

### Parallel pulls

```bash
# Pull different images in parallel (separate terminals)
lazyoci pull app1:latest --docker &
lazyoci pull app2:latest --docker &
lazyoci pull app3:latest --docker &
wait

echo "All pulls completed"
```

## Verify Your Setup

### Complete workflow test

```bash
# Pull a test image
lazyoci pull hello-world:latest --docker --quiet

# Verify it's in Docker
docker images hello-world

# Run the image  
docker run --rm hello-world

# Clean up
docker rmi hello-world:latest
```

If all steps complete successfully, your Docker integration is working correctly.

:::tip Performance Tip
Install skopeo for significantly faster Docker loading. Without skopeo, lazyoci converts the OCI format manually, which is slower but still functional.
:::