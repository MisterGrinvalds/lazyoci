---
sidebar_position: 2
title: First Browse
---

# Your First Browse Session

Now let's launch lazyoci and explore Docker Hub together. We'll search for nginx and examine its details.

## Launch the TUI

In your terminal, run lazyoci with no arguments:

```bash
lazyoci
```

You'll see the Terminal User Interface appear with four main panels:
- **Registry** (left): Shows available registries
- **Search** (top center): For finding repositories  
- **Artifacts** (bottom center): Lists found repositories/tags
- **Details** (right): Shows detailed information

You'll notice Docker Hub is already selected in the Registry panel.

## Navigate the Interface

Let's practice the basic navigation:

1. **Press `Tab`** to cycle between panels. Try it a few times to see the focus move around.

2. **Press `/`** to jump directly to the search box. You'll see the cursor appear in the search field.

## Search for nginx

With the search box focused:

1. **Type `nginx`** and press `Enter`

2. **Press `Tab`** to move focus to the Artifacts panel

You'll see a list of nginx-related repositories appear, with `nginx` at the top.

## Explore Repository Tags

1. **Press `Enter`** on the `nginx` repository

The Artifacts panel will now show nginx tags like `latest`, `alpine`, `stable`, etc.

2. **Use `j` and `k`** to move up and down through the tags

3. **Select `nginx:alpine`** and press `Enter`

## View Artifact Details

The Details panel now shows information about the nginx:alpine image:
- Digest and size information
- Creation date
- Architecture details
- Layer information

**Press `j` and `k`** to scroll through the details, or:
- **Press `g`** to jump to the top
- **Press `G`** to jump to the bottom

## Try Other Navigation Shortcuts

Experiment with these shortcuts:

- **Press `1`** to focus the Registry panel
- **Press `2`** to focus the Search panel  
- **Press `3`** to focus the Artifacts panel
- **Press `4`** to focus the Details panel
- **Press `?`** to see the help screen with all keybindings

## What You've Learned

You've successfully:
- ✅ Launched the lazyoci TUI
- ✅ Navigated between panels using Tab and number keys
- ✅ Searched for repositories on Docker Hub
- ✅ Browsed tags within a repository
- ✅ Viewed detailed artifact information

Next, let's [pull our first artifact](./first-pull.md)!