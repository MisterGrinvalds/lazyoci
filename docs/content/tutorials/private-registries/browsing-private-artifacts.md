---
sidebar_position: 3
title: Browsing Private Artifacts
---

# Browsing Private Artifacts

Now let's use the TUI to explore your private registry and its repositories. We'll see how private registries work alongside public ones.

## Launch the TUI

Start lazyoci:

```bash
lazyoci
```

## Select Your Private Registry

1. **Press `1`** to focus the Registry panel

2. **Use `j` and `k`** to navigate to your private registry

3. **Press `Enter`** to select it

You'll notice the registry panel now shows your private registry as active.

## Browse Private Repositories

1. **Press `2`** to focus the Search panel

2. **Leave the search empty** and press `Enter` to see all repositories

3. **Press `Tab`** to move to the Artifacts panel

You'll see a list of your organization's private repositories. These might include:
- Application images
- Base images
- Microservices
- Development tools

## Explore a Repository

1. **Select a repository** that interests you

2. **Press `Enter`** to view its tags

You'll see the available versions, including:
- Latest builds
- Version tags (like v1.2.3)
- Branch-based tags (like main, develop)

## View Private Artifact Details

1. **Select a specific tag**

2. **Press `4`** to focus the Details panel

The details show the same rich information as public images:
- Manifest digest and size
- Creation timestamp  
- Layer breakdown
- Architecture information

## Pull a Private Artifact

Let's pull one of your private artifacts:

1. **Select the artifact you want**

2. **Press `p`** to pull it

The authentication happens automatically using your stored credentials.

## Compare Registry Features

Try switching between your private registry and Docker Hub:

1. **Press `1`** to focus registries

2. **Select Docker Hub** and explore a few public repositories

3. **Switch back to your private registry**

Notice how the interface works consistently across both public and private registries.

## Search Private Repositories

Back in your private registry:

1. **Press `/`** to focus search

2. **Type part of an application name** your organization uses

3. **Press `Enter`** to search

You'll see filtered results showing only matching repositories.

## What You've Accomplished

You've successfully:
- ✅ Browsed your private registry in the TUI
- ✅ Explored private repositories and their tags
- ✅ Viewed detailed information about private artifacts
- ✅ Pulled private artifacts with automatic authentication
- ✅ Searched within your private registry
- ✅ Switched seamlessly between public and private registries

Great work! Now you're ready to explore [local development setups](../local-dev/).