---
sidebar_position: 1
title: Installation
---

# Installing lazyoci

Let's install lazyoci on your system. We'll show you two methods and verify the installation works.

## Method 1: Install with Go

If you have Go installed, this is the quickest method:

```bash
go install github.com/mistergrinvalds/lazyoci@latest
```

You should see Go downloading and building lazyoci. When it completes, the binary will be in your `$GOPATH/bin` directory.

## Method 2: Build from Source

First, clone the repository:

```bash
git clone https://github.com/mistergrinvalds/lazyoci.git
cd lazyoci
```

Now build the binary:

```bash
make build
```

You'll see output like:
```
go build -o bin/lazyoci ./cmd/lazyoci
```

The binary is now available at `bin/lazyoci`.

## Verify Installation

Let's confirm lazyoci is installed correctly:

```bash
lazyoci --help
```

You should see the help output with available commands:
```
lazyoci is a Terminal User Interface for browsing and managing OCI artifacts

Usage:
  lazyoci [command]

Available Commands:
  browse      Browse OCI artifacts and repositories
  config      Manage configuration
  pull        Pull OCI artifacts
  registry    Manage registry connections
  help        Help about any command
...
```

## What's Next

Great! Now that lazyoci is installed, let's [launch the TUI for the first time](./first-browse.md).