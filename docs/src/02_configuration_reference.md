# Configuration Reference

This section documents all configurable fields for a Cuestomize KRM function configuration.

## KRM Function Configuration

| Field          | Type   | Description                                                               |
| -------------- | ------ | ------------------------------------------------------------------------- |
| `apiVersion`   | string | API version. Unconstrained by default (CUE model can constrain it)        |
| `kind`         | string | Kind. Unconstrained by default (CUE model can constrain it)               |
| `metadata`     | object | Standard Kubernetes metadata*.                                            |
| `input`        | object | (Optional) Input sent to the model. Shape configured in the model itself. |
| `remoteModule` | object | (Optional) Remote CUE module configuration (OCI or CUE registry).         |
| `includes`     | object | (Optional) Additional resources to include in the CUE model.              |

### Metadata
The metadata field of the configuration must contain some annotations in order for `kustomize` to recognise it as a KRM function.
<br/>On top of that, Cuestomize offers some configurations options through the `.metadata` field.<br/>
All these options are documented below.

#### Annotations

`.metadata.annotations`

| Annotation                       | Description                                                                        |
| -------------------------------- | ---------------------------------------------------------------------------------- |
| `config.kubernetes.io/function`  | Contains the KRM function configuration.                                           |
| `config.cuestomize.io/validator` | If set to `"true"`, tells the function to use the CUE module for *validation* only |

##### Annotation – `config.kubernetes.io/function`
The annotation `config.kubernetes.io/function` is the one used by kustomize to configure a KRM function ([kustomize docs](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/containerized_krm_functions/#configuration)).

Its value must contains the configuration for the container that runs the KRM function.

```yaml
metadata:
  name: my-config
  annotations:
    config.kubernetes.io/function: |
      container:
        # the Cuestomize image you want to use
        image: ghcr.io/workday/cuestomize:latest

        # this is required to pull the CUE module from a registry
        network: true
```

> ⚠️ Passing environment variables to KRM functions is a discouraged practice (and may be removed in future kustomize versions), but is documented here for completeness. It also may be useful when developing to quickly iterate, without having to change the configuration.

The KRM function configuration also accepts environment variables to be passed to the container running the function, although that is discouraged and may be removed in future kustomize versions.

Cuestomize allows you to configure the logging level and pass the credentials for private registries through environment variables.

| Variable name       | Description                                       |
| ------------------- | ------------------------------------------------- |
| `LOG_LEVEL`         | The logging level (default: `warn`)               |
| `REGISTRY_USERNAME` | The registry to pull the CUE module from username |
| `REGISTRY_PASSWORD` | The registry to pull the CUE module from password |

##### Annotation – `config.cuestomize.io/validator`
Setting `config.cuestomize.io/validator: "true"` in the configuration annotations tells Cuestomize to use the CUE module as a validator only: it will unify the inputs and includes with the module, but it won't collect the outputs.

This is useful if you want to validate a set of manifests with some CUE constraints, e.g. ensuring that all Deployments use a particular `securityContext`, or that resources in certain namespaces has a particular label, etc.

When used in validator mode, CUE will be used to validate, instead of to generate, and the behaviour you can expect is the same as running [`cue eval` command](https://cuelang.org/docs/reference/command/cue-help-eval/).


### Example

```yaml
apiVersion: cuestomize.io/v1
kind: CuestomizeConfig
metadata:
  name: my-config
  annotations:
    config.kubernetes.io/function: |
      container:
        image: ghcr.io/workday/cuestomize:latest
        network: true
remoteModule:
  oci: docker.io/wackoninja/cuemodules:latest
includes:
- version: "v1"
  kind: ConfigMap
  name: "test-configmap"
  namespace: "test-namespace"
```
