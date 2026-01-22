## Module Pull From Private Registries (With Auth)

For private registries or repositories, you need to provide credentials. The recommended way is to use a Kubernetes Secret and reference it in your configuration.

### Secret Passing

This method requires you to have a _Kubernetes Secret_ in the _kustomize_ input stream having the credentials to access the private registry. The secret does not have to be present in the final output of kustomize, you can use the `config.kubernetes.io/local-config: "true"` annotation to tell kustomize to use the secret only during the build phase, and not show it anywhere in the rendered manifests. It is a standard way to pass sensitive data to KRM functions in kustomize.

You need to select a Kubernetes Secret through the `.remoteModule.auth` field:

```yaml
remoteModule:
  ref: ghcr.io/workday/cuestomize/cuemodules/cuestomize-examples-simple:latest
  auth:
    kind: Secret
    name: oci-auth
```

This tells Cuestomize to use the `oci-auth` Secret for authenticating to the registry.<br/>
The secret must be in the kustomize input stream to the function in order for it to be found and used by it.

> 💡 You can use Kustomize’s `secretGenerator` to create a Secret from environment variables:
>
> `.env` file
>
> ```env
> username=<username>
> password=<password>
> ```
>
> `kustomization.yaml
>
> ```yaml
> secretGenerator:
>   - name: oci-auth
>     envs:
>       - .env
>     options:
>       disableNameSuffixHash: true
>       annotations:
>         # ensures this secret is not included in the final output
>         config.kubernetes.io/local-config: "true"
> ```
>
> This will generate a Secret named `oci-auth` with your credentials.

#### Auth Secret Configuration

Several secret structures are supported, to accommodate different ways of storing credentials.

| Structure Type | Description                                                                                        |
| -------------- | -------------------------------------------------------------------------------------------------- |
| `Secret`       | Standard Kubernetes Secret with `username` and `password` fields in the `data` or `stringData`     |
| `DockerConfig` | Kubernetes Secret of type `kubernetes.io/dockerconfigjson`, containing a `.dockerconfigjson` field |

##### Structure Type – `Secret`

The `Secret` structure type expects a standard Kubernetes Secret containing the `username` and `password` fields in either the `data` or `stringData` sections.

The full list of supported fields is the following:

| Field          | Alternative Field        | Description                           |
| -------------- | ------------------------ | ------------------------------------- |
| `username`     | `REGISTRY_USERNAME`      | The registry username                 |
| `password`     | `REGISTRY_PASSWORD`      | The registry password                 |
| `accessToken`  | `REGISTRY_ACCESS_TOKEN`  | (Optional) The registry access token  |
| `refreshToken` | `REGISTRY_REFRESH_TOKEN` | (Optional) The registry refresh token |

##### Structure Type – `DockerConfig`

TODO: document

### Environment Variables (Discouraged)

This method of passing credentials is discouraged and may be removed in future kustomize versions, but is documented here for completeness, and because it may be useful when developing to quickly iterate.
