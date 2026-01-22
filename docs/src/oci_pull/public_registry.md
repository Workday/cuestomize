## Module Pull From Public Registries

> ⓘ No auth is required for public modules.

To pull from a public registry, you don't need to specify the `.remoteModule.auth` field to pass the credentials,
you just need to instruct the function on where the CUE module to pull is stored, and which tag you want to pull.

```yaml
apiVersion: cuestomize.dev/v1alpha1
kind: Cuestomization
metadata:
  name: example
  annotations:
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/function: |
      container:
        image: ghcr.io/workday/cuestomize:latest
        network: true
input:
  configMapName: example-configmap
remoteModule:
  ref: ghcr.io/workday/cuestomize/cuemodules/cuestomize-examples-simple:latest
```

| Field      | Description                                                       |
| ---------- | ----------------------------------------------------------------- |
| `ref`      | The full OCI reference in the format `registry/repo:tag`          |
| `registry` | (Deprecated) The OCI registry host (e.g., `ghcr.io`, `docker.io`) |
| `repo`     | (Deprecated) The repository path to your CUE module               |
| `tag`      | (Deprecated) The tag/version to pull                              |
