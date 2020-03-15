# Helm chart

Set env for the following:

```
export RELEASE_NAME=example
```

## Generate certs

Certs are required for webhook and part of the helm chart also needs to be filled first.
Before we install or render the Chart, generating certs must be done first.

```bash
./gen_certs.sh ${RELEASE_NAME}-trait-injector
```

It will fill certs-related placeholders in _values.yaml_ and have a backup of original file as _values.yaml.bak_

## Render deploy manifests

```bash
helm template ${RELEASE_NAME} .
```

## Configuration

Chart configuration are available in _values.yaml_ .
