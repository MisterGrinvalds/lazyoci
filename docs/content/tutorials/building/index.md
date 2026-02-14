---
sidebar_position: 1
title: Building and Pushing Artifacts
---

# Building and Pushing Artifacts

In this tutorial, you'll learn how to use `lazyoci build` to package and push different types of OCI artifacts from a `.lazy` configuration file.

## What you'll learn

- How to write a `.lazy` config file
- How to build and push a generic OCI artifact
- How to package and push a Helm chart
- How to use tag templates with git metadata

## Prerequisites

- lazyoci installed ([Installation guide](../getting-started/installation))
- Docker installed and running
- A local OCI registry running on `localhost:5050`

If you don't have a local registry, start one:

```bash
make registry-up
```

## Time estimate

10-15 minutes

## Steps

1. [Your First Build](./first-build) -- create a `.lazy` file and push a generic artifact
2. [Building a Helm Chart](./helm-chart) -- package and push a Helm chart
3. [Tag Templates](./tag-templates) -- use git metadata and variables in tags
