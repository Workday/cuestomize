# Getting Started

This section will guide you through the steps to get started with Cuestomize.

## The CUE Model

To get started with Cuestomize, you need to create a CUE module that defines your manifest generation logic.

You can either create your own CUE model from scratch, or use an existing one.

How to create CUE models that are compatible with Cuestomize is explained in different sections of this book, so we will use one of the existing models for this example.

## The Kustomization

You need to have a Kustomization directory that you can use to run Kustomize.
In this directory, you need to create a `kustomization.yaml` file that defines the resources you want to manage with Kustomize.

How to create a Kustomization project is out of the scope of this book, so we will assume you already have one, and the next steps will assume you are using the one under `examples/simple/kustomize`.

In your kustomization directory, you need to add a file that holds the configuration of the KRM function that will run Cuestomize. The name you give to this file is not important, as long as you reference it in the `transformers` section of your `kustomization.yaml` file.

### Create the KRM Function Configuration File

Change directory, and create a file named `krm-func.yaml` in the kustomization directory:

```bash
# We assume you have git cloned the repo locally and are in the root directory of the repo
cd examples/simple/kustomize

touch krm-func.yaml
```

> **Note:** the name you give to this file is not important, as long as you reference it in the `transformers` section of your `kustomization.yaml` file.

### Update the Kustomization File

Edit the `kustomization.yaml` file to add the `krm-func.yaml` file to the `transformers` section:

```yaml
kind: Kustomization

# ... other sections ...

transformers:
  - krm-func.yaml
```

## Configuring the KRM Function

Edit the `krm-func.yaml` file to configure the KRM function that will run Cuestomize.

Here is an example configuration:

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
includes:
  - group: apps
    version: v1
    kind: Deployment
    name: example-deployment
    namespace: example-namespace
  - version: v1
    kind: Service
    name: example-service
    namespace: example-namespace
remoteModule:
  ref: ghcr.io/workday/cuestomize/cuemodules/cuestomize-examples-simple:latest
```

> **Note:** Cuestomize does not constrain the `apiVersion` and `kind` fields of the KRM function configuration,
> so you can use whatever values you want, as long as they are valid Kubernetes resource names. In your CUE model,
> on the other hand, you can constrain these to specific values in order to ensure compatibility between the model
> and the function's configuration.

## Running Kustomize to Generate the Manifests

Now that you have everything set up, you can run Kustomize to generate the manifests.

Since Cuestomize is a KRM function, you'll need a few extra flags in order for `kustomize build` to work properly:

- `--enable-alpha-plugins` to enable the KRM function
- `--network` if your CUE model is pulled from a registry (can be omitted if the model is local to the function's image).

### Build the Manifests

```shell
kustomize build . --enable-alpha-plugins --network
```

You should see the build to be successful, and the CUE-generated manifests within the printed manifests.
