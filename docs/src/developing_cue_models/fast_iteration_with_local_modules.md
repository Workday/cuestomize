# Fast Iteration With Local Modules

Kustomize allows KRM functions to mount local directories as function modules. This feature can be leveraged during CUE module development to enable fast iteration, without needing to interact with an OCI registry or bake an image.

## Setting Up Local Module Mounting

To set up local module mounting, you need to create a kustomization file in the parent directory of your CUE module (kustomize constrains local mounts to be relative to the kustomization file, and you cannot go up in the directory tree).

Setup a directory structure like the following:

```
.
├── cue
│   └── cue.mod
│   └── main.cue
└── kustomization.yaml
└── krm-config.yaml
```

In this structure, the `cue` directory contains your CUE module files, while the `kustomization.yaml` file is in the parent directory.

Create a `kustomization.yaml` file with the following content:

```yaml
resources:
  - krm-config.yaml
transformers:
  - krm-config.yaml
```

In your KRM configuration file (`krm-config.yaml`), specify the function with a mount to the local CUE module directory:

```yaml
apiVersion: mymodule.cuestomize.dev/v1alpha1
kind: MyCUEFunction
metadata:
  name: example
  annotations:
    config.kubernetes.io/function: |
      container:
        image: ghcr.io/workday/cuestomize:latest
        mounts:
        - type: bind
          src: ./cue
          dst: /cue-resources # this is where the Cuestomize expects the CUE module to be
input: {} # your input configuration
```

## Running Cuestomize with Local Modules

With docker running (needed by kustomize to run containerised functions), you can now quickly test your CUE-Kustomize integration by running:

```bash
kustomize build . --enable-alpha-plugins
```

> Run the command in the directory where your `kustomization.yaml` file is located.
