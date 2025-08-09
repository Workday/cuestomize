package main

apiVersion: "cuestomize.dev/v1alpha1"
kind:       "Cuestomization"

input: {
	configMapName: "example-configmap"
}

includes: _

outConfigMap: {
	apiVersion: "v1"
	kind:       "ConfigMap"
	metadata: {
		name:      input.configMapName
		namespace: "default"
	}
	data: {
		deploymentName: includes["apps/v1"]["Deployment"]["example-namespace"]["example-deployment"].metadata.name
		serviceName:    includes["v1"]["Service"]["example-namespace"]["example-service"].metadata.name
	}
}

outputs: [
	outConfigMap,
]
