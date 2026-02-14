---
sidebar_position: 1
title: Installation
---

# Installing lazyoci

Let's install lazyoci on your system. Choose the method that fits your platform, then verify the installation works.

## macOS

### Homebrew (recommended)

```bash
brew install greenforests-studio/tap/lazyoci
```

Homebrew installs shell completions automatically.

### Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/mistergrinvalds/lazyoci/main/install.sh | sh
```

The script detects your OS and architecture, downloads the latest release, and installs the binary to `/usr/local/bin`. It also installs shell completions for bash, zsh, and fish if the standard directories exist.

To install without sudo:

```bash
curl -fsSL https://raw.githubusercontent.com/mistergrinvalds/lazyoci/main/install.sh | BINDIR=~/.local/bin sh
```

## Linux

### Homebrew

```bash
brew install greenforests-studio/tap/lazyoci
```

### Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/mistergrinvalds/lazyoci/main/install.sh | sh
```

### Debian / Ubuntu (from GitHub release)

Download the `.tar.gz` archive from the [releases page](https://github.com/mistergrinvalds/lazyoci/releases) and extract:

```bash
# Download (replace VERSION and ARCH as needed)
curl -fsSL https://github.com/mistergrinvalds/lazyoci/releases/download/v0.1.0/lazyoci_0.1.0_linux_amd64.tar.gz -o lazyoci.tar.gz
tar -xzf lazyoci.tar.gz
sudo install -m 755 lazyoci /usr/local/bin/lazyoci
```

## Windows

### Scoop (recommended)

```powershell
scoop bucket add greenforests https://github.com/greenforests-studio/scoop-bucket
scoop install lazyoci
```

### Manual Download

Download the `.zip` archive for Windows from the [releases page](https://github.com/mistergrinvalds/lazyoci/releases) and add the extracted directory to your `PATH`.

## Go Install

If you have Go 1.23+ installed:

```bash
go install github.com/mistergrinvalds/lazyoci/cmd/lazyoci@latest
```

The binary will be in `$GOPATH/bin`. This method does not install shell completions -- see [Shell Completions](#shell-completions) below.

## From Source

```bash
git clone https://github.com/mistergrinvalds/lazyoci.git
cd lazyoci
make install              # installs to /usr/local/bin
make install-completions  # installs shell completions
```

To install to a different prefix:

```bash
make install PREFIX=~/.local
```

## Shell Completions

Homebrew, Scoop, and the install script handle completions automatically. For `go install` or other methods, set up completions manually:

### Bash

Add to your `~/.bashrc`:

```bash
source <(lazyoci completion bash)
```

### Zsh

Add to your `~/.zshrc`:

```bash
source <(lazyoci completion zsh)
```

Or generate the file once:

```bash
lazyoci completion zsh > "${fpath[1]}/_lazyoci"
```

### Fish

```bash
lazyoci completion fish | source
```

Or save permanently:

```bash
lazyoci completion fish > ~/.config/fish/completions/lazyoci.fish
```

### PowerShell

Add to your `$PROFILE`:

```powershell
lazyoci completion powershell | Out-String | Invoke-Expression
```

## Verify Installation

Confirm lazyoci is installed correctly:

```bash
lazyoci version
```

You should see output like:

```
lazyoci version 0.1.0 (commit: abc1234, built: 2026-01-15T10:00:00Z)
```

Check the help to see available commands:

```bash
lazyoci --help
```

```
lazyoci is a Terminal User Interface for browsing and managing OCI artifacts

Usage:
  lazyoci [command]

Available Commands:
  browse      Browse OCI artifacts and repositories
  build       Build and push OCI artifacts
  config      Manage configuration
  pull        Pull OCI artifacts
  registry    Manage registry connections
  version     Print version information
  help        Help about any command
...
```

## What's Next

Now that lazyoci is installed, let's [launch the TUI for the first time](./first-browse.md).
