# Validator Mode

Validator mode allows you to use CUE to validate your Kubernetes resources, instead of generating them.

In this mode, you can write CUE schemas that define the expected structure and constraints of your Kubernetes resources, and then use the `includes` field in your KRM function configuration to specify which resources should be validated against those schemas.

For example, you can define a CUE schema for Kubernetes workloads resources (e.g., Pods, DaemonSets, Deployments, etc.) that enforces certain security best practices, and then use that schema to validate your Kubernetes resources at render time, ensuring that they meet your security requirements before they are applied to the cluster.

To enable validator mode, add the following annotation to your KRM function configuration:

```yaml
config.cuestomize.io/validator: "true"
```

An example of a KRM function configuration that enables validator mode and includes a CUE schema for validating Kubernetes workloads resources can be found in the [examples/validation/kustomize/validator.yaml](https://github.com/Workday/cuestomize/blob/main/examples/validation/kustomize/validator.yaml).

The following is an example of a KRM function configuration that enables validator mode and forwards all `v1`, `apps/v1`, and `batch/v1` resources to the CUE module for validation:

```yaml
apiVersion: cuestomize.dev/v1alpha1
kind: Validator
metadata:
  name: example
  annotations:
    config.cuestomize.io/validator: "true"
~    config.kubernetes.io/function: |
~      container:
~        image: ghcr.io/workday/cuestomize:latest
~        network: true
includes:
  - group: ""
    version: v1
  - group: apps
    version: v1
  - group: batch
    version: v1
remoteModule:
  ref: github.com/workday/cuestomize/cuemodules/cuestomize-examples-validation:latest
```

## Validator CUE Module

A CUE module that can be used for validator mode does not require to have an `outputs` field, although it may still have it defined (it will be ignored).
