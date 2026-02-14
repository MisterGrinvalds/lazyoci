---
sidebar_position: 2
title: Building a Helm Chart
---

# Building a Helm Chart

Let's package a Helm chart as an OCI artifact and push it to a registry. No Helm CLI is required -- lazyoci handles the packaging directly.

## Create a chart directory

```bash
mkdir -p ~/lazyoci-tutorial/mychart/templates
```

## Create Chart.yaml

```bash
cat > ~/lazyoci-tutorial/mychart/Chart.yaml << 'EOF'
apiVersion: v2
name: mychart
version: 0.1.0
description: My first OCI Helm chart
type: application
appVersion: "1.0.0"
EOF
```

## Create a template

```bash
cat > ~/lazyoci-tutorial/mychart/templates/configmap.yaml << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-config
data:
  greeting: "hello from lazyoci"
EOF
```

## Create values.yaml

```bash
cat > ~/lazyoci-tutorial/mychart/values.yaml << 'EOF'
replicaCount: 1
EOF
```

## Update the .lazy config

Replace the `.lazy` file to include the Helm chart:

```bash
cat > ~/lazyoci-tutorial/.lazy << 'EOF'
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
          - v1.0.0
          - latest

  - type: helm
    name: mychart
    chartPath: mychart
    targets:
      - registry: localhost:5050/tutorial/charts/mychart
        tags:
          - "{{ .ChartVersion }}"
          - latest
EOF
```

## Preview

```bash
cd ~/lazyoci-tutorial
lazyoci build --dry-run
```

Notice how `{{ .ChartVersion }}` resolves to `0.1.0` from `Chart.yaml`:

```
Building mychart (type: helm)...
  [dry-run] would push localhost:5050/tutorial/charts/mychart:0.1.0
  [dry-run] would push localhost:5050/tutorial/charts/mychart:latest
```

## Build only the chart

Use `--artifact` to build just the Helm chart:

```bash
lazyoci build --artifact mychart --insecure
```

## Verify

```bash
lazyoci browse tags localhost:5050/tutorial/charts/mychart
```

You should see `0.1.0` and `latest`.

## What you learned

- `type: helm` packages a chart directory as an OCI artifact
- `chartPath` points to the directory containing `Chart.yaml`
- `{{ .ChartVersion }}` is automatically read from `Chart.yaml`
- `--artifact` filters to a specific artifact in a multi-artifact config

Next, let's explore [tag templates](./tag-templates).
