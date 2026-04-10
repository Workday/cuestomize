# Includes

In this section, the Cuestomize _includes_ mechanism is explained, and some use cases on when it can be useful are provided.

The _includes_ mechanism allows you to _"forward"_ resources from the kustomize input stream to the CUE module, either to be used as additional input data (e.g. inferring the namespace in which to deploy resources) or for validation purposes.

## How to Forward Resources to the CUE Module

To forward resources from the kustomize input stream to your CUE module, you need to add an [`includes`](../02_configuration_reference.md#KRM-Function-Configuration) field to your KRM function configuration. This field accepts a list of selectors that specify which resources to include.

Each selector can filter resources by:

- `group` and `version` (the Kubernetes API group and version)
- `kind` (the Kubernetes resource kind)
- `namespace` (the Kubernetes namespace, regex supported)
- `name` (the name of the resource, regex supported)
- `labelSelector` (label-based selection)
- `annotationSelector` (annotation-based selection)

Here's an example:

```yaml
apiVersion: cuestomize.dev/v1alpha1
kind: Cuestomization
metadata:
  name: example
~  annotations:
~    config.kubernetes.io/function: |
~      container:
~        image: ghcr.io/workday/cuestomize:latest
~        network: true
input:
  configMapName: my-config
includes:
  # Forward all Services from v1
  - version: v1
    kind: Service
  # Forward all Deployments from apps/v1 with label app=myapp
  - group: apps
    version: v1
    kind: Deployment
    labelSelector: "app=myapp"
  # Forward all ConfigMaps with specific annotation
  - group: "" # group can be omitted for core resources
    version: v1
    kind: ConfigMap
    annotationSelector: "config-type=template"
  # Forward a specific Pod by name using regex
  - group: ""
    version: v1
    kind: Pod
    name: "my-pod-.*"
```

All resources matching any of the selectors will be included in the `includes` structure and made available to your CUE module.

## How to Access Included Resources in the CUE Module

Resources forwarded through the `includes` field are made available in the CUE module through a structured `includes` map. This map is organized hierarchically as:

```cue
includes: <apiVersion>: <kind>: <namespace>: <name>: {resource}
```

For example, with the configuration above, you can access included resources like this:

```cue
package main

includes: _

// Access a specific Service
myService: includes["v1"]["Service"]["default"]["my-service"]

// Access a specific Deployment
myDeployment: includes["apps/v1"]["Deployment"]["my-namespace"]["my-deployment"]

// Access all Services
allServices: includes["v1"]["Service"]

// Access all resources of a kind in a namespace
allDeploymentsInNamespace: includes["apps/v1"]["Deployment"]["production"]
```

In your CUE constraints or outputs, you can then use these resources. For example:

```cue
// Extract the namespace from an included resource
deploymentNamespace: includes["apps/v1"]["Deployment"]["my-namespace"]["my-deployment"].metadata.namespace

// Use data from an included resource
serviceName: includes["v1"]["Service"]["default"]["my-service"].metadata.name

// Iterate over multiple resources
serviceNames: [for name, svc in includes["v1"]["Service"]["default"] {
  svc.metadata.name
}]
```

## Use Cases

The includes mechanism is useful in several scenarios, like when you want to:

- Reference existing resources in the input stream to inform your generated resources (e.g., infer the namespace from an included Namespace resource)
- Validate resources in the input stream against constraints defined in your CUE module (e.g., ensure all included Deployments have certain labels or resource limits).

### Includes as Additional Input Data

Using includes as additional input data allows your CUE module to reference other resources in the kustomize input stream and extract information from them. This is helpful when you need your generated resources to be aware of or coordinated with existing resources.

#### Example – Inferring Namespace

In this example, the CUE module uses an included Namespace to determine where to deploy resources:

KRM function configuration:

```yaml
apiVersion: cuestomize.dev/v1alpha1
kind: NginxDeployment
metadata:
  name: nginx-deployment
~  annotations:
~    config.kubernetes.io/function: |
~      container:
~        image: ghcr.io/workday/cuestomize:latest
~        network: true
input:
  deploymentName: nginx-deployment
  image: nginx:latest
  replicas: 3
includes:
  - version: v1
    kind: Namespace
    labelSelector: "environment=production"
```

CUE module:

```cue
package main

apiVersion: "cuestomize.dev/v1alpha1"
kind: "NginxDeployment"

input: {
  deploymentName!: string
  image!: string
  replicas!: int
}

// helper to extract a single resource from includes, with error handling
_#ExtractNs: {
  i: [string]: [string]: [string]: [string]: [string]: _
	_resources: list.FlattenN([
		for k, rs in i["v1"]["Namespace"][""] {[
			for kk, r in rs {r},
		]},
	], 1)

	if len(_resources) != 1 {
		error("Expected 1 resource of kind \(i.kind) in namespace \(i.namespace), but found \(len(_resources))")
	}
	out: _resources[0]
}


includes: {} | null

_namespaces: includes["v1"]["Namespace"][""]

// Extract the namespace from the included Namespace resource
_targetNamespace: {_#ExtractNs & {i: includes}}.out

outDeployment: {
  apiVersion: "apps/v1"
  kind: "Deployment"
  metadata: {
    name: input.deploymentName
    namespace: _targetNamespace  // Use the namespace from the included resource
  }
  spec: {
    replicas: input.replicas
    selector: {
      matchLabels: {
        app: input.deploymentName
      }
    }
    template: {
      metadata: {
        labels: {
          app: input.deploymentName
        }
      }
      spec: {
        containers: [{
          name: "nginx"
          image: input.image
        }]
      }
    }
  }
}

outputs: deployment: outDeployment
```

#### Example – Synchronizing Metadata

Your CUE module can also synchronize labels and annotations from existing resources:

```cue
package main

input: {
  appName!: string
}

includes: {}

// inherit labels from an included resource. (This example assumes you know the namespace name ahead of time
// , but you could also extract it like in the previous example)
_inheritedLabels: includes["v1"]["Namespace"][""]["example-ns"].metadata.labels

outService: {
  apiVersion: "v1"
  kind: "Service"
  metadata: {
    name: input.appName
    labels: _inheritedLabels  // Use labels from the included Namespace
  }
  spec: {
    ports: [{
      port: 80
      targetPort: 8080
    }]
    selector: {
      app: input.appName
    }
  }
}

outputs: service: outService
```

### Includes for Validation

Using includes for validation mode allows you to enforce constraints on resources in the kustomize input stream. This is powerful for ensuring that all resources meet certain requirements before they are applied to the cluster.

When using Cuestomize in validator mode (by setting the `config.cuestomize.io/validator: "true"` annotation), included resources are validated against your CUE constraints, and any violations will cause the build to fail.

#### Example – Validating Resource Constraints

In this example, the CUE validator ensures that all included Deployments follow certain security and configuration standards:

KRM function configuration:

```yaml
apiVersion: cuestomize.dev/v1alpha1
kind: Validator
metadata:
  name: deployment-validator
  annotations:
    config.cuestomize.io/validator: "true"
~    config.kubernetes.io/function: |
~      container:
~        image: ghcr.io/workday/cuestomize:latest
~        network: true
includes:
  - group: apps
    version: v1
    kind: Deployment
```

CUE validator module:

```cue
package main

#DeploymentConstraints: {
  // ... define your constraints here
}

includes: {
  // ensure all included Deployments meet the constraints
  "apps/v1": "Deployment": [_]: [_]: #DeploymentConstraints
}
```

When running kustomize build, if any Deployment in the input stream violates these constraints (e.g., missing labels, exceeding resource limits, not setting security constraints), the build will fail with a validation error.
