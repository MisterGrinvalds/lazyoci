---
sidebar_position: 2
title: Pushing Artifacts
---

# Pushing Test Artifacts

Let's populate our local registry with different types of OCI artifacts. This will give us interesting content to explore in lazyoci.

## Prerequisites Check

Make sure you're in the lazyoci source directory:

```bash
pwd
```

You should see something like `/path/to/lazyoci`. If not, clone and navigate to the repository first.

Also verify your local registry is still running:

```bash
docker ps | grep registry
```

You should see the registry container running on port 5050.

## Push All Test Fixtures

The lazyoci project includes a convenient make target that pushes various test artifacts:

```bash
make registry-push-all
```

You'll see output as different artifact types are built and pushed:
```
Pushing container image...
Pushing Helm chart...
Pushing WASM module...
Pushing Cosign signature...
Pushing SBOM...
✓ All test fixtures pushed to localhost:5050
```

This creates a variety of artifacts that showcase different OCI media types.

## Verify the Push

Let's check what repositories now exist:

```bash
lazyoci browse repos localhost:5050
```

You'll see several test repositories:
```
test/app
test/helm-chart  
test/wasm-module
test/signatures
test/sbom
```

## Browse a Specific Repository

Let's look at the tags in one repository:

```bash
lazyoci browse tags localhost:5050/test/app
```

You should see tags like:
```
latest
v1.0.0
dev
```

## Check Repository Counts

Get a summary of what we pushed:

```bash
lazyoci browse repos localhost:5050 --output json | jq length
```

This shows how many repositories are now in your local registry.

## Manual Push (Alternative)

If you want to push a simple container image manually:

```bash
# Pull a small image
docker pull alpine:latest

# Tag it for your local registry
docker tag alpine:latest localhost:5050/test/alpine:latest

# Push it
docker push localhost:5050/test/alpine:latest
```

## What You've Created

Your local registry now contains:
- ✅ Multiple test repositories with different artifact types
- ✅ Container images with various tags
- ✅ Helm charts, WASM modules, and other OCI artifacts
- ✅ A rich testing environment for lazyoci features

Perfect! Now let's [explore these artifact types](./artifact-types.md) in the TUI.