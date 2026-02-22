---
title: CLI Commands
---

# CLI Commands

Command tree for the lazyoci CLI interface.

## Root Command

```
lazyoci
```

Launches TUI when no subcommand specified.

**Persistent Flags:**
- `--output`, `-o` (default: `text`, values: `text`, `json`, `yaml`)
- `--artifact-dir` (default: `""`)
- `--theme` (default: `""`)

## Command Tree

```
lazyoci
├── pull <reference>
├── build [path]
├── mirror
├── browse
│   ├── repos <registry-url>
│   ├── tags <registry/repo>
│   ├── manifest <registry/repo:tag>
│   └── search <registry> <query>
├── registry
│   ├── list
│   ├── add <url>
│   ├── remove <url>
│   └── test <url>
├── config
│   ├── get <key>
│   ├── set <key> <value>
│   ├── list
│   └── path
├── version
└── completion
    ├── bash
    ├── zsh
    ├── fish
    └── powershell
```

## Command Arguments

| Command | Arguments | Type |
|---------|-----------|------|
| `pull` | `<reference>` | ExactArgs(1) |
| `build` | `[path]` | MaximumNArgs(1) |
| `mirror` | (none) | NoArgs |
| `browse repos` | `<registry-url>` | ExactArgs(1) |
| `browse tags` | `<registry/repo>` | ExactArgs(1) |
| `browse manifest` | `<registry/repo:tag>` | ExactArgs(1) |
| `browse search` | `<registry> <query>` | ExactArgs(2) |
| `registry add` | `<url>` | ExactArgs(1) |
| `registry remove` | `<url>` | ExactArgs(1) |
| `registry test` | `<url>` | ExactArgs(1) |
| `config get` | `<key>` | ExactArgs(1) |
| `config set` | `<key> <value>` | ExactArgs(2) |
| `version` | (none) | NoArgs |
| `completion bash` | (none) | NoArgs |
| `completion zsh` | (none) | NoArgs |
| `completion fish` | (none) | NoArgs |
| `completion powershell` | (none) | NoArgs |