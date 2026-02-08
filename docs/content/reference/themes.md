---
title: Themes
---

# Themes

Available themes for the lazyoci TUI interface.

## Available Themes

| Theme ID | Display Name | Type |
|----------|--------------|------|
| `default` | Default | Light/Dark |
| `catppuccin-latte` | Catppuccin Latte | Light |
| `catppuccin-mocha` | Catppuccin Mocha | Dark |
| `dracula` | Dracula | Dark |
| `gruvbox` | Gruvbox | Light/Dark |
| `solarized-dark` | Solarized Dark | Dark |
| `tokyonight` | Tokyo Night | Dark |

**Total:** 7 themes

## Theme Selection

### Configuration File

Set in `config.yaml`:

```yaml
theme: "catppuccin-mocha"
```

### Command Line

```bash
lazyoci --theme dracula
```

### TUI Interface

Press `T` to open theme picker while in TUI.

## Theme Variants

### Light/Dark Support

Themes support automatic light/dark mode detection:

| Theme | Light Variant | Dark Variant |
|-------|---------------|--------------|
| `default` | ✓ | ✓ |
| `gruvbox` | ✓ | ✓ |

### Dark-only Themes

| Theme | Description |
|-------|-------------|
| `catppuccin-mocha` | Warm dark theme with purple accents |
| `dracula` | Dark theme with vibrant colors |
| `solarized-dark` | Low-contrast dark theme |
| `tokyonight` | Dark theme inspired by Tokyo at night |

### Light-only Themes

| Theme | Description |
|-------|-------------|
| `catppuccin-latte` | Light theme with warm colors |

## Color Mode Configuration

Control light/dark mode behavior:

```yaml
mode: "auto"    # Detect from terminal
mode: "dark"    # Force dark mode
mode: "light"   # Force light mode
```

## Theme Detection

### Automatic Detection

When `mode: "auto"` (default), lazyoci detects terminal background using:

1. `COLORFGBG` environment variable
2. Terminal capability queries
3. Fallback to dark mode

### COLORFGBG Format

```bash
export COLORFGBG="15;0"   # Light foreground, dark background
export COLORFGBG="0;15"   # Dark foreground, light background
```

**Background values:**
- `0-6` → Dark mode
- `7-15` → Light mode

## Default Behavior

Without explicit configuration:

1. Theme defaults to `"default"`
2. Mode defaults to `"auto"`
3. Terminal background is detected automatically
4. Appropriate theme variant is selected