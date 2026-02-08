---
title: TUI Keybindings
---

# TUI Keybindings

Complete keybinding reference for the lazyoci TUI interface.

## Global Keybindings

Active when not in input mode.

| Key | Action |
|-----|--------|
| `Tab` | Cycle focus forward |
| `Shift+Tab` | Focus registry list |
| `Ctrl+C` | Quit application |
| `q` | Quit application |
| `?` | Show help |
| `1` | Focus registry list |
| `2` | Focus search |
| `3` | Focus artifacts |
| `4` | Focus details |
| `/` | Focus search input |
| `S` | Open settings |
| `T` | Open theme picker |

## Input Mode Keybindings

Active when focused on input fields.

| Key | Action |
|-----|--------|
| `Escape` | Exit input, focus registry |
| `Ctrl+C` | Quit application |
| `Tab` | Cycle focus forward |

## View-Specific Keybindings

Active in specific views regardless of input mode.

| Key | Action | Views |
|-----|--------|-------|
| `j` | Move down | All lists |
| `k` | Move up | All lists |
| `g` | Go to top | Details view |
| `G` | Go to bottom | Details view |
| `Enter` | Select item | All lists |
| `p` | Pull artifact | Artifact lists |
| `d` | Pull to Docker | Artifact lists |

## Focus Cycle

The focus cycle determines navigation order when pressing `Tab`:

1. **Registry List** - Available registries
2. **Search Input** - Search query field  
3. **Artifact Filter** - Artifact filtering
4. **Details** - Artifact details view
5. **Registry List** - (cycles back)

## Context-Specific Behavior

### Registry List
- `Enter` - Select registry and load repositories
- `j`/`k` - Navigate registry list

### Search Input  
- Text input for search queries
- `Enter` - Execute search
- `Escape` - Clear focus, return to registry list

### Artifact Filter
- Filter artifacts in current view
- Standard list navigation applies

### Details View
- `g`/`G` - Navigate to top/bottom of content
- `j`/`k` - Scroll content
- `p` - Pull current artifact
- `d` - Pull current artifact to Docker