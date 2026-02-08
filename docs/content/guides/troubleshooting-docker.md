---
title: Fix Docker Load Failures
description: Solve problems loading artifacts into Docker
sidebar_position: 10
---

# Fix Docker Load Failures

Diagnose and resolve issues when loading pulled artifacts into Docker. These problems typically occur after successful artifact downloads.

## Identify Docker Load Issues

### Common error messages

- `docker daemon not running`
- `permission denied while connecting to docker`
- `failed to load image into docker`
- `skopeo not found, falling back to manual conversion`
- `invalid OCI layout`

### Isolate the problem

```bash
# Test pull without Docker loading
lazyoci pull nginx:latest --dest /tmp/test-pull --quiet

# If this succeeds, the issue is with Docker loading, not pulling
```

## Check Docker Daemon

### Verify Docker is running

```bash
# Check Docker daemon status
docker info
```

If this fails, Docker daemon is not running or accessible.

### Start Docker daemon

**Linux (systemd):**
```bash
# Start Docker service
sudo systemctl start docker

# Enable auto-start
sudo systemctl enable docker
```

**macOS:**
```bash
# Start Docker Desktop
open /Applications/Docker.app
```

**Windows:**
Start Docker Desktop from the Start menu.

### Test Docker access

```bash
# Test basic Docker functionality
docker run --rm hello-world
```

## Fix Permission Issues

### Check Docker socket permissions

```bash
# Check Docker socket
ls -la /var/run/docker.sock
```

Should be accessible to your user (either owned by you or group-accessible).

### Add user to docker group

```bash
# Add current user to docker group
sudo usermod -aG docker $USER

# Apply group changes (logout/login or use newgrp)
newgrp docker
```

### Test permission fix

```bash
# Test Docker access without sudo
docker ps
```

Should work without permission errors.

## Install and Configure Skopeo

Skopeo significantly improves Docker loading performance.

### Install skopeo

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

**Arch Linux:**
```bash
sudo pacman -S skopeo
```

### Verify skopeo installation

```bash
# Check skopeo version
skopeo --version

# Test skopeo Docker access
skopeo inspect docker-daemon:hello-world:latest
```

### Test skopeo Docker loading

```bash
# Pull artifact to directory
lazyoci pull alpine:latest --dest /tmp/alpine-test

# Manually load with skopeo
skopeo copy oci:/tmp/alpine-test docker-daemon:alpine:test

# Verify in Docker
docker images alpine
```

## Troubleshoot Manual OCI Conversion

When skopeo is not available, lazyoci converts OCI to Docker format manually.

### Check OCI layout

```bash
# Pull artifact and inspect layout
lazyoci pull nginx:latest --dest /tmp/nginx-oci

# Check OCI layout structure
ls -la /tmp/nginx-oci/
cat /tmp/nginx-oci/oci-layout
cat /tmp/nginx-oci/index.json
```

Should contain valid OCI layout files.

### Verify blob integrity

```bash
# Check blobs directory
ls -la /tmp/nginx-oci/blobs/sha256/

# Verify blob checksums (example)
cd /tmp/nginx-oci/blobs/sha256/
for blob in *; do
  echo "Checking $blob..."
  echo "$blob  $blob" | sha256sum -c
done
```

All checksums should verify correctly.

### Test manual Docker load

```bash
# Try manual Docker load (requires working OCI to Docker conversion)
# This is done internally by lazyoci, but you can test the Docker load directly
docker load < /path/to/docker-save-format.tar
```

## Fix Docker API Issues

### Check Docker API version

```bash
# Check Docker API version
docker version
```

Look for API version compatibility between client and daemon.

### Test Docker API directly

```bash
# Test Docker API endpoint
curl --unix-socket /var/run/docker.sock http://localhost/version
```

Should return JSON with Docker version information.

### Fix API compatibility

```bash
# Set specific Docker API version if needed
export DOCKER_API_VERSION=1.41

# Test Docker commands
docker ps
```

## Debug Storage Issues

### Check disk space

```bash
# Check available disk space
df -h /var/lib/docker  # Linux
df -h ~/Library/Containers/com.docker.docker  # macOS
```

Docker loading can fail if there's insufficient disk space.

### Clean up Docker

```bash
# Remove unused images
docker image prune -f

# Remove unused containers
docker container prune -f

# Full system cleanup
docker system prune -f
```

### Check Docker root directory

```bash
# Check Docker storage location
docker info | grep "Docker Root Dir"
```

Ensure this location has sufficient space and proper permissions.

## Test Docker Integration

### Test basic Docker loading

```bash
# Pull small image with Docker loading
lazyoci pull hello-world:latest --docker --quiet

# Verify in Docker
docker images hello-world
```

### Test with different image types

```bash
# Test with different images
lazyoci pull alpine:latest --docker --quiet
lazyoci pull nginx:latest --docker --quiet
lazyoci pull postgres:15 --docker --quiet

# Verify all loaded successfully
docker images
```

### Test platform-specific loading

```bash
# Test cross-platform loading
lazyoci pull --platform linux/arm64 alpine:latest --docker --quiet

# Verify platform in Docker
docker inspect alpine:latest | grep Architecture
```

## Alternative Docker Loading Methods

### Use docker import

If standard loading fails:

```bash
# Pull to directory
lazyoci pull alpine:latest --dest /tmp/alpine-oci

# Export as tar and import (manual process)
# This requires custom scripting to convert OCI to tar
```

### Use docker save/load workflow

```bash
# If you have the image elsewhere
docker save alpine:latest | gzip > alpine.tar.gz
docker load < alpine.tar.gz
```

### Use buildkit import

```bash
# With Docker Buildx
docker buildx imagetools create --tag alpine:imported alpine:latest
```

## Verify Docker Loading Fix

### Test complete workflow

```bash
# Test full pull and load
lazyoci pull nginx:latest --docker --quiet

# Verify image is available
docker images nginx

# Test running the image
docker run --rm nginx:latest nginx -v

# Clean up
docker rmi nginx:latest
```

### Test with private registry

```bash
# Test with authenticated registry
lazyoci pull your-registry.com/your-repo:tag --docker --quiet

# Verify private image loaded
docker images your-registry.com/your-repo
```

### Test performance

```bash
# Time the Docker loading process
time lazyoci pull large-image:latest --docker --quiet
```

With skopeo installed, this should be significantly faster than without.

## Monitor Docker Resources

### Check Docker daemon logs

**Linux:**
```bash
# Check Docker daemon logs
sudo journalctl -u docker.service -f
```

**macOS:**
```bash
# Check Docker Desktop logs
~/Library/Containers/com.docker.docker/Data/log/vm/docker.log
```

### Monitor Docker during loading

```bash
# Monitor Docker in separate terminal
watch docker images

# Run pull in another terminal
lazyoci pull large-image:latest --docker
```

You should see the image appear in the Docker images list after loading completes.

:::tip Docker Troubleshooting
Start by verifying Docker daemon is running and accessible with `docker info`. Most Docker loading issues stem from daemon connectivity or permission problems rather than lazyoci-specific issues.
:::