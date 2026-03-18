package main

import (
	v "validation.cuestomize.dev/pkg/validations"
)

input: _

includes: {
	"apps/v1": {
		"DaemonSet": [_]: [_]:   v.#STIGDaemonSet
		"Deployment": [_]: [_]:  v.#STIGDeployment
		"StatefulSet": [_]: [_]: v.#STIGStatefulSet
	}
	"batch/v1": {
		"CronJob": [_]: [_]: v.#STIGCronJob
		"Job": [_]: [_]:     v.#STIGJob
	}
	"v1": {
		"Pod": [_]: [_]: v.#STIGPod
	}
}

outputs: _
outputs: {}
