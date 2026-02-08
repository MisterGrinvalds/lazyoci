---
title: Configure Custom Storage
description: Change where artifacts are cached locally
sidebar_position: 7
---

# Configure Custom Storage

Control where lazyoci stores downloaded artifacts by configuring the artifact storage location. Useful for managing disk space, performance optimization, and organizational requirements.

## Understanding Storage Hierarchy

lazyoci determines the artifact storage location using this priority order:

1. **Command-line flag** - `--artifact-dir` on specific commands
2. **Environment variable** - `LAZYOCI_ARTIFACT_DIR`
3. **Configuration file** - `artifactDir` in config
4. **Default location** - `~/.cache/lazyoci/artifacts/`

## Set Storage via Configuration

Configure the storage location permanently using the config command.

### Set custom directory

```bash
lazyoci config set artifact-dir ~/my-artifacts --create
```

The `--create` flag creates the directory if it doesn't exist.

### Verify configuration

```bash
lazyoci config show
```

Look for the `artifactDir` setting in the output.

### Test custom storage

```bash
lazyoci pull alpine:latest --quiet
ls -la ~/my-artifacts/oci/
```

You should see the alpine artifact stored in your custom location.

## Set Storage via Environment Variable

Configure storage temporarily or for specific environments.

### Set environment variable

```bash
export LAZYOCI_ARTIFACT_DIR="/fast-ssd/lazyoci-cache"
```

### Create the directory

```bash
mkdir -p "$LAZYOCI_ARTIFACT_DIR"
```

### Pull with custom storage

```bash
lazyoci pull nginx:latest --quiet
ls -la "$LAZYOCI_ARTIFACT_DIR"
```

### Make permanent

Add to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
echo 'export LAZYOCI_ARTIFACT_DIR="/fast-ssd/lazyoci-cache"' >> ~/.bashrc
source ~/.bashrc
```

## Set Storage via Command Flag

Override storage location for specific pull operations.

### Pull to specific directory

```bash
lazyoci pull redis:latest --artifact-dir /tmp/redis-artifacts
```

This stores only this pull in the specified location.

### Pull multiple images to same location

```bash
# Pull several images to shared location
lazyoci pull postgres:15 --artifact-dir /shared/db-images
lazyoci pull mysql:8.0 --artifact-dir /shared/db-images
lazyoci pull redis:latest --artifact-dir /shared/db-images

# Verify all are stored together
ls -la /shared/db-images/
```

## Performance Optimization

### Use fast storage

Point to SSD or fast network storage for better performance:

```bash
# Configure to use SSD mount
lazyoci config set artifact-dir /mnt/fast-ssd/lazyoci --create
```

### Use local storage in containers

When running in containers, use local storage:

```bash
# In Dockerfile or container environment
ENV LAZYOCI_ARTIFACT_DIR=/tmp/lazyoci-cache
```

### Network storage for shared access

Configure network storage for team sharing:

```bash
# Configure to use NFS mount
export LAZYOCI_ARTIFACT_DIR="/mnt/nfs-share/lazyoci-artifacts"
mkdir -p "$LAZYOCI_ARTIFACT_DIR"
```

## Storage Organization

### Understand storage structure

The artifact directory contains this structure:

```
artifacts/
├── oci/
│   ├── registry.example.com/
│   │   └── repo/
│   │       └── tag/
│   │           ├── blobs/
│   │           ├── index.json
│   │           └── oci-layout
└── docker/
    └── ... (if Docker integration is used)
```

### Separate by environment

```bash
# Development environment
export LAZYOCI_ARTIFACT_DIR="$HOME/.cache/lazyoci-dev"

# Production environment  
export LAZYOCI_ARTIFACT_DIR="/opt/lazyoci-prod"

# Staging environment
export LAZYOCI_ARTIFACT_DIR="/opt/lazyoci-staging"
```

### Separate by project

```bash
# Project-specific storage
cd /path/to/project
export LAZYOCI_ARTIFACT_DIR="$(pwd)/.lazyoci-cache"
```

## Disk Space Management

### Check storage usage

```bash
# Check current usage
du -sh ~/.cache/lazyoci/artifacts/
# or your custom directory
du -sh /path/to/custom/artifacts/
```

### Clean up old artifacts

```bash
# Remove artifacts older than 30 days
find ~/.cache/lazyoci/artifacts/ -type f -mtime +30 -delete
find ~/.cache/lazyoci/artifacts/ -type d -empty -delete
```

### Move existing artifacts

```bash
# Move from default to custom location
mkdir -p /new/location
mv ~/.cache/lazyoci/artifacts/* /new/location/
lazyoci config set artifact-dir /new/location
```

## Temporary Storage Scenarios

### One-time pulls to temporary location

```bash
# Pull to /tmp for temporary use
lazyoci pull large-image:latest --artifact-dir /tmp/one-time-pull --dest /tmp/extracted
```

### CI/CD with ephemeral storage

```bash
#!/bin/bash
# CI script with temporary cache
export LAZYOCI_ARTIFACT_DIR="/tmp/ci-cache"
mkdir -p "$LAZYOCI_ARTIFACT_DIR"

# Pull required images
lazyoci pull app:${BUILD_TAG} --docker --quiet
lazyoci pull tests:${BUILD_TAG} --docker --quiet

# Cache is automatically cleaned up when CI completes
```

## Troubleshoot Storage Issues

### Check effective storage location

```bash
# Pull with verbose output to see where files go
lazyoci pull hello-world:latest
```

Look for messages indicating where the artifact is being stored.

### Verify directory permissions

```bash
# Check if directory is writable
touch "$(lazyoci config show | grep artifactDir | awk '{print $2}')/test"
rm "$(lazyoci config show | grep artifactDir | awk '{print $2}')/test"
```

### Check available space

```bash
# Check free space in storage directory
df -h ~/.cache/lazyoci/artifacts/
# or your custom directory
df -h /path/to/custom/artifacts/
```

### Reset to default

```bash
# Remove custom configuration
lazyoci config unset artifact-dir

# Clear environment variable
unset LAZYOCI_ARTIFACT_DIR

# Verify default is used
lazyoci pull hello-world:latest --quiet
ls -la ~/.cache/lazyoci/artifacts/
```

## Verify Your Configuration

### Test storage hierarchy

```bash
# 1. Set config file location
lazyoci config set artifact-dir ~/config-storage --create

# 2. Set environment variable (higher priority)
export LAZYOCI_ARTIFACT_DIR=~/env-storage
mkdir -p ~/env-storage

# 3. Use command flag (highest priority)
lazyoci pull alpine:latest --artifact-dir ~/flag-storage

# Verify artifact went to flag-storage (highest priority)
ls -la ~/flag-storage/
```

### Confirm configuration persistence

```bash
# Pull without flags (should use config or env)
lazyoci pull nginx:latest --quiet

# Check where it was stored
ls -la ~/env-storage/  # Should contain nginx if env var is set
```

The storage configuration is working correctly when artifacts appear in the expected location based on the priority hierarchy.

:::tip Storage Strategy
Use the configuration file for permanent settings, environment variables for environment-specific overrides, and command flags for one-off storage locations. This provides flexibility while maintaining predictable behavior.
:::