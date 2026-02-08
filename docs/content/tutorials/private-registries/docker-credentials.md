---
sidebar_position: 2
title: Docker Credentials
---

# Using Existing Docker Credentials

If you've already logged into registries with `docker login`, lazyoci can use those credentials automatically. Let's see how this works.

## Check Your Docker Config

First, let's see what's in your Docker configuration:

```bash
cat ~/.docker/config.json
```

You might see something like this:
```json
{
  "auths": {
    "https://index.docker.io/v1/": {},
    "my-company-registry.com": {}
  },
  "credsStore": "desktop"
}
```

Notice the empty `{}` objects in `auths` - the actual credentials are stored in the credential helper (`credsStore`).

## Test with Docker Hub

If you've done `docker login` before, try pulling from Docker Hub without adding credentials:

```bash
lazyoci pull my-private-repo/my-app:latest
```

lazyoci will automatically find and use your Docker credentials!

## Login to a New Registry via Docker

Let's add a registry through Docker first:

```bash
docker login my-company-registry.com
```

Enter your username and password when prompted.

## Test Automatic Detection

Now try using that registry in lazyoci without explicitly adding it:

```bash
lazyoci browse repos my-company-registry.com
```

lazyoci found your credentials automatically! You'll see your private repositories listed.

## Add to lazyoci Registry List

To make the registry appear in the TUI registry panel, add it without credentials:

```bash
lazyoci registry add https://my-company-registry.com --name "My Company"
```

Since lazyoci can already access your Docker credentials, you don't need to specify `--user` and `--pass`.

## How It Works

lazyoci checks for credentials in this order:

1. **Explicitly added credentials** (via `registry add --user --pass`)
2. **Docker credential helpers** (from `~/.docker/config.json`)
3. **Docker config auths** (for legacy stored credentials)

This means your existing Docker workflow continues to work seamlessly.

## What You've Learned

You've discovered:
- ✅ lazyoci automatically uses Docker credentials
- ✅ How Docker stores credentials with credential helpers
- ✅ You can add registries without re-entering passwords
- ✅ The credential lookup order lazyoci uses

Ready to [browse your private artifacts](./browsing-private-artifacts.md)?