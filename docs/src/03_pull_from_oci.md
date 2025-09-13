# Pull From OCI Registry

Cuestomize can fetch CUE modules directly from an OCI registry (such as Docker Hub or GitHub Container Registry). This allows you to share and reuse CUE logic across projects and teams, and makes distributing your CUE models easier.

To pull a CUE module from an OCI registry, specify the `remoteModule` field in your Cuestomize KRM configuration.

## Public Registries

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
  registry: ghcr.io
  repo: workday/cuestomize/cuemodules/cuestomize-examples-simple
  tag: latest
```

| Field      | Description                                          |
| ---------- | ---------------------------------------------------- |
| `registry` | The OCI registry host (e.g., `ghcr.io`, `docker.io`) |
| `repo`     | The repository path to your CUE module               |
| `tag`      | The tag/version to pull                              |

## Private Registries (With Auth)

For private registries or repositories, you need to provide credentials. The recommended way is to use a Kubernetes Secret and reference it in your configuration.

You need to select a Kubernetes Secret through the `remoteModule.auth` field:

```yaml
remoteModule:
  registry: ghcr.io
  repo: workday/cuestomize/cuemodules/cuestomize-examples-simple
  tag: latest
  auth:
    kind: Secret
    name: oci-auth
```

This tells Cuestomize to use the `oci-auth` Secret for authenticating to the registry.<br/>
The secret must be in the kustomize input stream to the function in order for it to be found and used by it.

> 💡 You can use Kustomize’s `secretGenerator` to create a Secret from environment variables:
> 
> `.env` file
> ```env
> username=<username>
> password=<password>
> ```
> 
> `kustomization.yaml
> ```yaml
> secretGenerator:
> - name: oci-auth
>   envs:
>   - .env
>   options:
>     disableNameSuffixHash: true
>     annotations:
>       config.kubernetes.io/local-config: "true"
> ```
> 
> This will generate a Secret named `oci-auth` with your credentials.
