# Cuestomize
Cuestomize is a Kubernetes Package Manager using CUE-lang and integrated in Kustomize.

It is implemented as a Kustomize KRM function that reads a CUE model, and optionally some input resources from the Kustomize stream, and passes back to Kustomize the generated resources.

It provides the type-safety of CUE and the flexibility of kustomize, combined in a single tool.
Moreover, it allows your CUE model to consume resources from the Kustomize stream, which can be used to feed the CUE model with input data (as well as the `input` section of the KRM function's specification).

The CUE model can then use the input values and resources to generate the output manifests.

The CUE model can either be pulled from an OCI registry, or be local to the KRM function (in which case you need to package a Docker image with both the CUE model and the Cuestomize binary).

## Usage
If you have a compatible CUE model already, you can use from kustomize as follows (look at the `example` directory for more information):
- Add it to the `transformers` section of your Kustomization file
  ```yaml
  transformers:
  - krm-func.yaml
  ```
- Then configure the KRM function in the `krm-func.yaml` file (or any name you gave to it)
  ```yaml
  apiVersion: cuestomize.dev/v1alpha1 # or whatever apiVersion your CUE model expects
  kind: Cuestomization # or whatever kind your CUE model expects
  metadata:
    name: my-cuestomization
    annotations:
      config.kubernetes.io/function: |
        container:
          image: ghcr.io/workday/cuestomize:latest
          network: true
  input:
    replicas: 3
    createRBAC: true
  includes:
    - kind: Namespace
      name: my-namespace
  ```

Make sure to pass `kustomize build` the following flags:
- `--enable-alpha-plugins` to enable the KRM function
- `--network` if your CUE model is pulled from a registry.

> **Example:** `kustomize build . --enable-alpha-plugins --network .`.

## CUE Model Integration
Cuestomize is able to integrate with any CUE model respecting the following constraints:
- The model accepts a `input` section (you are free to decide the structure of this section to match the expected KRM input structure)
- The model has an `outputs` section which is a slice of KRM resources. This field will hold the generated resources
- The model (optionally) accepts an `includes` section which is a map `<apiVersion>:<kind>:<namespace>:<name>:{resource}` of resources that are forwarded from the kustomize input stream to the CUE model.
