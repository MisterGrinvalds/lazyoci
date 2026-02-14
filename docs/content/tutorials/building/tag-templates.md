---
sidebar_position: 3
title: Tag Templates
---

# Tag Templates

Tag values in `.lazy` files support Go template variables that are resolved at build time. This lets you automatically include git metadata, timestamps, and CLI-provided values in your tags.

## Available variables

| Variable | Source |
|----------|--------|
| `{{ .Tag }}` | `--tag` CLI flag |
| `{{ .GitSHA }}` | Short git commit SHA |
| `{{ .GitBranch }}` | Current git branch |
| `{{ .ChartVersion }}` | `Chart.yaml` version (helm only) |
| `{{ .Timestamp }}` | UTC time as `YYYYMMDDHHmmss` |

## Try it out

Make sure you're in a git repository:

```bash
cd ~/lazyoci-tutorial
git init && git add -A && git commit -m "initial"
```

## Update tags to use templates

```bash
cat > .lazy << 'EOF'
version: 1
artifacts:
  - type: artifact
    name: app-config
    mediaType: application/vnd.example.config.v1
    files:
      - path: config.json
        mediaType: application/json
    targets:
      - registry: localhost:5050/tutorial/app-config
        tags:
          - "{{ .Tag }}"
          - "{{ .GitSHA }}"
          - "{{ .Tag }}-{{ .GitBranch }}"
          - latest
EOF
```

## Preview the resolved tags

```bash
lazyoci build --tag v2.0.0 --dry-run
```

Output:

```
Building app-config (type: artifact)...
  [dry-run] would push localhost:5050/tutorial/app-config:v2.0.0
  [dry-run] would push localhost:5050/tutorial/app-config:a1b2c3d
  [dry-run] would push localhost:5050/tutorial/app-config:v2.0.0-main
  [dry-run] would push localhost:5050/tutorial/app-config:latest
```

## Combine variables

Templates can be combined in a single tag:

```yaml
tags:
  - "{{ .Tag }}-{{ .GitSHA }}"       # v2.0.0-a1b2c3d
  - "{{ .GitBranch }}-{{ .Timestamp }}"  # main-20260209153000
```

## Literal tags

Tags without `{{ }}` are used as-is:

```yaml
tags:
  - latest        # always "latest"
  - stable        # always "stable"
  - "{{ .Tag }}"  # from --tag flag
```

## What you learned

- Tag templates use Go `{{ }}` syntax
- `--tag` sets the `{{ .Tag }}` variable
- Git metadata is resolved automatically from the local repo
- Variables can be combined in a single tag string
- Tags without template delimiters are literal values

## Clean up

```bash
rm -rf ~/lazyoci-tutorial
```

## Next steps

- See the [`.lazy` Config Reference](/reference/lazy-config) for the full schema
- Read the [Building Artifacts Guide](/guides/building-artifacts) for more patterns
- Check the [examples/](https://github.com/mistergrinvalds/lazyoci/tree/main/examples) directory for complete working examples
