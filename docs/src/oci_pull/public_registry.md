## Module Pull From Public Registries

> ⓘ No auth is required for public modules.

To pull from a public registry, you just need to configure the function on where the CUE module to pull is stored, and which tag you want to pull.
You can do this by specifying a `remoteModule` field in your Cuestomization resource, like so:

```yaml
{{#include ../../../examples/simple/kustomize/krm-func.yaml}}
```

In this example, we are pulling the `ghcr.io/workday/cuestomize/cuemodules/cuestomize-examples-simple` module at the `latest` tag.
Just by specifying `.remoteModule.ref`, Cuestomize will try to pull the module from the public registry before doing its processing.

Obviously, the OCI reference pulled must be a valid CUE module in order for Cuestomize to process it correctly.
