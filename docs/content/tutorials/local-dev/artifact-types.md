---
sidebar_position: 3
title: Artifact Types
---

# Exploring Different Artifact Types

Now let's explore the various OCI artifact types in your local registry. lazyoci uses color coding and badges to help you identify different artifact types at a glance.

## Launch the TUI

Start lazyoci and connect to your local registry:

```bash
lazyoci
```

1. **Press `1`** to focus the Registry panel
2. **Select "Local Dev"** using `j`/`k` and `Enter`

## Browse All Repositories

1. **Press `2`** to focus the Search panel
2. **Leave search empty** and press `Enter` to see all repos
3. **Press `Tab`** to move to the Artifacts panel

You'll see your test repositories listed with different visual indicators.

## Examine Container Images

1. **Select `test/app`** (or similar container image repo)
2. **Press `Enter`** to view tags

Look for the **blue "IMG" badge** - this indicates container images. These artifacts can be run as containers.

3. **Press `4`** to view details of a selected tag

The details panel shows:
- Layer information (each layer of the container filesystem)
- Size and architecture
- Creation timestamp

## Explore Helm Charts

1. **Go back** (press `3` then navigate to a Helm chart repository)
2. **Look for the green "HELM" badge**

Helm charts are packaged Kubernetes applications stored as OCI artifacts.

3. **View the details** to see the chart metadata and files included

## Check WASM Modules

1. **Navigate to a WASM repository**
2. **Look for the purple "WASM" badge**

WebAssembly modules are compiled code that can run in various environments, now distributed as OCI artifacts.

## View Signatures and Attestations

1. **Find repositories with signatures**
2. **Look for the orange "SIG" badge**

These are cryptographic signatures created by tools like Cosign to verify artifact authenticity.

## Examine SBOMs

1. **Navigate to SBOM repositories**
2. **Look for the yellow "SBOM" badge**

Software Bills of Materials (SBOMs) list all components and dependencies in an artifact.

## Compare Artifact Details

Try selecting different artifact types and comparing their details:

- **Container images** show layers and filesystem information
- **Helm charts** show chart metadata and template files
- **WASM modules** show module metadata and entry points
- **Signatures** show cryptographic verification data
- **SBOMs** show dependency trees and component lists

## Use Different View Options

While exploring, try these TUI features:

- **Press `T`** to cycle through different color themes
- **Press `S`** to access settings and customize the view
- **Press `j`/`k`** to scroll through long details
- **Press `g`/`G`** to jump to top/bottom of details

## Pull Different Artifact Types

Try pulling various artifacts:

1. **Select a container image** and press `p`
2. **Select a Helm chart** and press `p`
3. **Select a WASM module** and press `p`

Notice how lazyoci handles each artifact type appropriately.

## What You've Discovered

You now understand:
- ✅ How lazyoci visually distinguishes artifact types with color-coded badges
- ✅ The different kinds of OCI artifacts beyond just container images
- ✅ How each artifact type shows relevant metadata in the details panel
- ✅ That the same TUI interface works consistently across all artifact types
- ✅ How modern registries store diverse artifact types, not just containers

Congratulations! You've completed all the lazyoci tutorials and now have a solid foundation for using the tool effectively. You can explore more advanced features in the [How-to Guides](/guides/) or dive deeper into concepts in the [Reference](/reference/cli/) section.