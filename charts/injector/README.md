# Helm chart

## Generate deploy manifests

```bash
helm template $release-name .
```

## Webhook Configuration

options:
- rules, e.g. Deployment/StatefulSet Create
- clientConfig, e.g. caBundle
- namespaceSelector
- objectSelector
