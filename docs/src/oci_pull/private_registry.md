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

The following structure types are supported for the auth secret:

| Structure Type | Description                                                                                    |
| -------------- | ---------------------------------------------------------------------------------------------- |
| `Secret`       | Standard Kubernetes Secret with `username` and `password` fields in the `data` or `stringData` |

> Support may be expanded in the future to include other types, such as Docker config files.

##### Structure Type – `Secret`

The `Secret` structure type expects a standard Kubernetes Secret containing the `username` and `password` fields in either the `data` or `stringData` sections.

The full list of supported fields is the following:

| Field          | Alternative Field        | Description                           |
| -------------- | ------------------------ | ------------------------------------- |
| `username`     | `REGISTRY_USERNAME`      | The registry username                 |
| `password`     | `REGISTRY_PASSWORD`      | The registry password                 |
| `accessToken`  | `REGISTRY_ACCESS_TOKEN`  | (Optional) The registry access token  |
| `refreshToken` | `REGISTRY_REFRESH_TOKEN` | (Optional) The registry refresh token |

### Environment Variables (Discouraged)

> This method of passing credentials is discouraged and may be removed in future kustomize versions, but is documented here for completeness, and because it may be useful when developing to quickly iterate.

Kustomize lets you pass environment variables to KRM functions, and Cuestomize supports passing the registry credentials through the same environment variables it would expect in the Secret structure.

An example on how to pass the credentials through environment variables:

```yaml
apiVersion: cuestomize.dev/v1alpha1
kind: Cuestomization
metadata:
  name: example
  annotations:
    config.kubernetes.io/function: |
      container:
        image: ghcr.io/workday/cuestomize:latest
        network: true
        envs: [REGISTRY_USERNAME, REGISTRY_PASSWORD]
```

In this case, you would need to set the `REGISTRY_USERNAME` and `REGISTRY_PASSWORD` environment variables in your shell before running `kustomize build`:

```bash
export REGISTRY_USERNAME=<username>
export REGISTRY_PASSWORD=<password>
kustomize build .
```

This allows you not to have the credentials stored in any Kubernetes Secret, and avoid having to include them in the kustomize function configuration.

Another way, even more discouraged, is to hardcode the credentials directly in the configuration:

```yaml
apiVersion: cuestomize.dev/v1alpha1
kind: Cuestomization
metadata:
  name: example
  annotations:
    config.kubernetes.io/function: |
      container:
        image: ghcr.io/workday/cuestomize:latest
        network: true
        envs:
        - REGISTRY_USERNAME="<username>"
        - REGISTRY_PASSWORD="<password>"
```

This is not recommended at all, as it exposes the credentials in plain text in your configuration files, which may be committed to version control or shared with others.

Please avoid using this method as much as possible, and prefer the Secret passing method described above.
