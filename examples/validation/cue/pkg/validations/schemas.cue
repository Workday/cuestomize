package validations

import (
	api "k8s.io/api/core/v1"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
)

// #STIGPod is a Kubernetes Pod respecting STIG requirements
#STIGPod: api.#Pod & {
	spec: #STIGPodSpec
}

// #STIGDeployment is a Kubernetes Deployment respecting STIG requirements
#STIGDeployment: apps.#Deployment & {
	spec: template: spec: #STIGPodSpec
}

// #STIGStatefulSet is a Kubernetes StatefulSet respecting STIG requirements
#STIGStatefulSet: apps.#StatefulSet & {
	spec: template: spec: #STIGPodSpec
}

// #STIGDaemonSet is a Kubernetes DaemonSet respecting STIG requirements
#STIGDaemonSet: apps.#DaemonSet & {
	spec: template: spec: #STIGPodSpec
}

// #STIGCronJob is a Kubernetes CronJob respecting STIG requirements
#STIGCronJob: batch.#CronJob & {
	spec: jobTemplate: spec: template: spec: #STIGPodSpec
}

// #STIGJob is a Kubernetes Job respecting STIG requirements
#STIGJob: batch.#Job & {
	spec: template: spec: #STIGPodSpec
}
