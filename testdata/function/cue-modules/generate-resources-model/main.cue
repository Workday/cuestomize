package main

import (
	api "k8s.io/api/core/v1"
)

input: {}

includes: {
	"v1": "Namespace": "": _: api.#Namespace
}

// helper to extract a single namespace resource from includes, with error handling
_#ExtractNamespace: {
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

inputNs: _#ExtractNamespace & {i: includes}

editedNs: inputNs & {
	metadata: {
		annotations: {
			"edited": "true"
		}
	}
}

outputs: {
	ns: editedNs
	ksa: {
		apiVersion: "v1"
		kind:       "ServiceAccount"
		metadata: {
			name:      "test-serviceaccount"
			namespace: editedNs.metadata.name
		}
	}
	deploy: {
		apiVersion: "apps/v1"
		kind:       "Deployment"
		metadata: {
			name:      "test-deployment"
			namespace: editedNs.metadata.name
		}
		spec: {
			serviceAccountName: ksa.metadata.name
			replicas:           1
			selector: {
				matchLabels: {app: "test-app"}
			}
			template: {
				metadata: {
					labels: {app: "test-app"}
				}
				spec: {
					containers: [{
						name:  "test-container"
						image: "nginx:latest"
					}]
				}
			}
		}
	}
}
