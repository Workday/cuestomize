# CUE Model Structure

When developing CUE modules that integrate with Cuestomize, it is important to first understand how Cuestomize interacts with CUE, and which
structures are required in order to ensure compatibility.

| Structure    | Sourced From                                                                                  | Purpose                                                                                   |
| ------------ | --------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| `apiVersion` | The KRM configuration `apiVersion` field.                                                     | Allows CUE module developers to constrain the supported version of the KRM configuration. |
| `kind`       | The KRM configuration `kind` field.                                                           | Allows to constrain the supported kind of the KRM configuration.                          |
| `metadata`   | The KRM configuration `metadata` field.                                                       | Allows to access metadata information from the KRM configuration.                         |
| `input`      | The KRM configuration `input` field.                                                          | Access the inputs from the KRM configuration.                                             |
| `includes`   | The resources in the _kustomize_ input stream matching the configuration `includes` selectors | Access the included resources from the KRM configuration.                                 |

Here is an example of what a CUE model would see from a given KRM configuration:

```yaml
apiVersion: cuestomize.dev/v1alpha1
kind: Cuestomization
metadata:
  name: example
input:
  name: example
  annotations:
    example-annotation: example-value
  replicas: 3
includes:
  - version: v1
    kind: Namespace
    annotationSelector: app=example
```

> The configuration above selects all Namespaces with the label `app=example` from the input stream, and includes them
> in the `includes` structure of the CUE module.

The above KRM configuration would provide the following structures to the CUE module (assuming there is a matching Namespace in the input stream):

```cue
apiVersion: "cuestomize.dev/v1alpha1"
kind: "Cuestomization"
metadata: {
    name: string
}

input: {
    name: example
    annotations: {
        example-annotation: example-value
    }
    replicas: 3
}

// includes structure: <apiVersion>: <kind>: <namespace>: <name>: {object}
includes: {
    "v1": {
        "Namespace": {
            "": {
                "example-namespace": {
                    metadata: {
                        name: "example-namespace"
                        annotations: {
                            "app": "example"
                        }
                    }
                }
            }
        }
    }
}
```
