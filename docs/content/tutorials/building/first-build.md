---
sidebar_position: 1
title: Your First Build
---

# Your First Build

Let's create a `.lazy` config and push a generic OCI artifact to a local registry.

## Create a project directory

```bash
mkdir ~/lazyoci-tutorial && cd ~/lazyoci-tutorial
```

## Create a file to push

Create a simple JSON config file that we'll push as an OCI artifact:

```bash
cat > config.json << 'EOF'
{
  "app": "hello",
  "version": "1.0.0",
  "environment": "production"
}
EOF
```

## Write the .lazy config

Create a `.lazy` file that describes how to package and push this file:

```bash
cat > .lazy << 'EOF'
version: 1
artifacts:
  - type: artifact
    name: app-config
    mediaType: application/vnd.example.config.v1
    files:
      - path: config.json
        mediaType: application/json
    targets:
      - registry: localhost:5050/tutorial/app-config
        tags:
          - v1.0.0
          - latest
EOF
```

## Preview with dry run

Before actually building and pushing, preview what will happen:

```bash
lazyoci build --dry-run
```

You should see:

```
Building from /Users/you/lazyoci-tutorial/.lazy...
Mode: dry-run
Building app-config (type: artifact)...
  [dry-run] would push localhost:5050/tutorial/app-config:v1.0.0
  [dry-run] would push localhost:5050/tutorial/app-config:latest

OK    app-config (artifact)
      localhost:5050/tutorial/app-config:v1.0.0 [built]
      localhost:5050/tutorial/app-config:latest [built]
```

## Build and push

Now push for real:

```bash
lazyoci build --insecure
```

The `--insecure` flag is needed because the local registry uses HTTP.

## Verify the push

Check that the artifact exists in the registry:

```bash
lazyoci browse tags localhost:5050/tutorial/app-config
```

You should see `v1.0.0` and `latest` listed.

## Inspect the manifest

```bash
lazyoci browse manifest localhost:5050/tutorial/app-config:v1.0.0 -o json
```

You'll see the OCI manifest with your config.json as a layer.

## Pull it back

```bash
lazyoci pull localhost:5050/tutorial/app-config:v1.0.0 --insecure
```

## What you learned

- A `.lazy` file defines artifacts to build and push
- `type: artifact` packages arbitrary files as OCI artifacts
- `--dry-run` previews the operation without side effects
- `--insecure` allows pushing to HTTP registries

Next, let's [build a Helm chart](./helm-chart).
