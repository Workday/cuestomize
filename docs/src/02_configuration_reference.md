# Configuration Reference

This section documents all configurable fields for a Cuestomize KRM function configuration.

## KRM Function Configuration

| Field          | Type   | Description                                                               |
| -------------- | ------ | ------------------------------------------------------------------------- |
| `apiVersion`   | string | API version. Unconstrained by default (CUE model can constrain it)        |
| `kind`         | string | Kind. Unconstrained by default (CUE model can constrain it)               |
| `metadata`     | object | Standard Kubernetes metadata.                                             |
| `input`        | object | (Optional) Input sent to the model. Shape configured in the model itself. |
| `includes`     | object | (Optional) Additional resources to include in the CUE model.              |
| `remoteModule` | object | (Optional) Remote CUE module configuration (OCI or CUE registry).         |

### Metadata

The metadata field of the configuration must contain some annotations in order for `kustomize` to recognise it as a KRM function.
<br/>On top of that, Cuestomize offers some configurations options through the `.metadata` field.<br/>
All these options are documented below.

#### Annotations

`.metadata.annotations`

| Annotation                       | Description                                                                        |
| -------------------------------- | ---------------------------------------------------------------------------------- |
| `config.kubernetes.io/function`  | Contains the KRM function configuration.                                           |
| `config.cuestomize.io/validator` | If set to `"true"`, tells the function to use the CUE module for _validation_ only |

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

### Input

Input is an `object` whose shape depends on the CUE model you are integrating with.

The CUE model creator (yourself, or a third party) defines the shape of the input directly in the model itself.

The allowed shape and values for the `input` are entirely dependent on the CUE model constraints the model imposes.

For example, if the CUE model defines the input as:

```cue
input: {
    name: string
    replicas: int | *1
    region: "us-west-1" | "us-east-1" | "eu-central-1"
}
```

the `input` section of the KRM function configuration must conform to that shape:

```yaml
# valid input section
input:
  name: my-app
  replicas: 3
  region: us-west-1
```

On the other hand, the following input would be invalid, and Cuestomize would raise an error at runtime:

```yaml
# invalid input section
input:
  name: my-app
  replicas: "three" # invalid: defined as integer in the CUE model
  region: ap-southeast-1 # invalid: not one of the allowed values
```

As you can see, the `input` section is entirely dependent on the CUE model, if you change the model, the shape of the input must be updated accordingly.

### Includes

The `includes` field is the one that ties with the concept of _includes_ in Cuestomize.
Includes, in Cuestomize, are an advanced feature that allows you to forward resources from the Kustomize input stream to the CUE model, to be used as additional input.

The power lies in the ability to let CUE "infer" some values from the kustomize stream, without forcing the user to pass them explicitly through the `input` section. For example, you may want to forward the Namespace resource from the kustomize stream to the CUE model, so that the model can use the received Namespace's name to set the namespace of the generated resources, or to read and use some labels or annotations from it.

The `includes` field is a list of resource selectors, and resources matching one of the selectors will be forwarded to the CUE model in the `includes` field.

### Remote Module

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
  ref: docker.io/wackoninja/cuemodules:latest
includes:
  - version: "v1"
    kind: ConfigMap
    name: "test-configmap"
    namespace: "test-namespace"
```
