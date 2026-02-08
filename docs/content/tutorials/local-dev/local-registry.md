---
sidebar_position: 1
title: Local Registry
---

# Starting Your Local Registry

Let's set up a local OCI registry for development. This gives you a safe place to experiment without affecting production registries.

## Start the Registry

If you have the lazyoci source code, there's a pre-configured Docker Compose file:

```bash
docker compose -f docker-compose.dev.yml up -d
```

You'll see output like:
```
[+] Running 2/2
 ✔ Network lazyoci_default  Created
 ✔ Container registry       Started
```

If you don't have the source code, create a simple registry:

```bash
docker run -d -p 5050:5000 --name local-registry registry:2
```

## Verify the Registry

Check that the registry is running:

```bash
curl http://localhost:5050/v2/
```

You should see:
```json
{}
```

This empty JSON response means the registry is alive and responding.

## Add to lazyoci

Now let's add this local registry to lazyoci:

```bash
lazyoci registry add http://localhost:5050 \
  --name "Local Dev" \
  --insecure
```

We use `--insecure` because our local registry doesn't have TLS certificates.

You'll see confirmation:
```
✓ Added registry http://localhost:5050 as "Local Dev"
```

## Test the Connection

Verify lazyoci can connect:

```bash
lazyoci registry test http://localhost:5050
```

You should see:
```
✓ Connection to http://localhost:5050 successful
✓ No authentication required
```

## Browse the Empty Registry

Let's see what an empty registry looks like:

```bash
lazyoci browse repos localhost:5050
```

You'll get output showing no repositories found:
```
No repositories found
```

This is expected - we just started with a fresh registry!

## Check in the TUI

Launch lazyoci and select your local registry:

```bash
lazyoci
```

1. **Press `1`** to focus the Registry panel
2. **Navigate to "Local Dev"** using `j` and `k`
3. **Press `Enter`** to select it

The registry is connected and ready, but empty.

## What You've Set Up

You now have:
- ✅ A local OCI registry running on port 5050
- ✅ The registry added to lazyoci as "Local Dev"
- ✅ A verified connection to the local registry
- ✅ An empty registry ready for test artifacts

Next, let's [push some test artifacts](./pushing-artifacts.md) to make it interesting!