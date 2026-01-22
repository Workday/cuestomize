## Module Pull From Private Registries (With Auth)

For private registries or repositories, you need to provide credentials. The recommended way is to use a Kubernetes Secret and reference it in your configuration.

You need to select a Kubernetes Secret through the `remoteModule.auth` field:

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
