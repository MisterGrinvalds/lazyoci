---
sidebar_position: 3
title: First Pull
---

# Pulling Your First Artifact

Let's pull the nginx:alpine image we found in the previous step. We'll do this both through the TUI and the CLI.

## Pull via TUI

If you still have lazyoci open with nginx:alpine selected:

1. **Press `p`** to pull the artifact

You'll see a progress indication as lazyoci downloads the image layers. When complete, you'll see the artifact saved to your local cache directory.

If you closed lazyoci, reopen it and navigate back to nginx:alpine, then press `p`.

## Pull via CLI

Let's also try pulling using the command line. **Press `q`** to quit the TUI first.

Now run this command:

```bash
lazyoci pull nginx:alpine
```

You'll see output like:
```
Pulling nginx:alpine...
✓ Pulled nginx:alpine to ~/.cache/lazyoci/artifacts/docker.io/library/nginx/alpine/...
```

## Pull Directly to Docker

We can also pull an image directly to your local Docker daemon:

```bash
lazyoci pull nginx:alpine --docker
```

You'll see:
```
Pulling nginx:alpine...
✓ Pulled nginx:alpine to Docker
```

Verify it worked:

```bash
docker images | grep nginx
```

You should see the nginx image listed:
```
nginx        alpine    abc123...    2 days ago    41MB
```

## Explore CLI Browse Commands

The CLI also has powerful browse capabilities. Try these:

**List repositories on Docker Hub:**
```bash
lazyoci browse repos docker.io
```

**Search for repositories:**
```bash
lazyoci browse search docker.io nginx
```

**Browse tags for nginx:**
```bash
lazyoci browse tags nginx --limit 5
```

**View manifest details:**
```bash
lazyoci browse manifest nginx:alpine
```

You'll see detailed JSON output for each command.

## Specify Different Platforms

You can pull images for specific platforms:

```bash
lazyoci pull nginx:alpine --platform linux/arm64
```

## What You've Accomplished

Congratulations! You've successfully:
- ✅ Pulled an artifact through the TUI using the `p` key
- ✅ Pulled an artifact via CLI using `lazyoci pull`
- ✅ Pulled directly to Docker with the `--docker` flag
- ✅ Verified the image in your Docker daemon
- ✅ Explored CLI browse commands
- ✅ Learned about platform-specific pulls

You now know the basics of lazyoci! Ready to explore [private registries](../private-registries/)?