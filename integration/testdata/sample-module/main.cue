package main

input: {}

cm: {
	apiVersion: "v1"
	kind:       "ConfigMap"
	metadata: {
		name:      "sample-cm"
		namespace: "default"
	}
	data: {
		"key1": "value1"
		"key2": "value2"
	}
}

outputs: [
	cm,
]
