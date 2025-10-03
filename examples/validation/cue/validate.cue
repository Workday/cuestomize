package main

import "cue.k8s.example/deployment"

includes: _

deployments: includes["apps/v1"]["Deployment"]["example-namespace"]

#Labels: [string]: string

#Deployment: {
	apiVersion: "apps/v1"
	kind:       "Deployment"

	metadata!: {
		// assert every deployment's name must start with "example-"
		name!: =~"example-.*"

		// must have labels as a map of strings
		labels!: #Labels

		...
	}

	spec!: {
		// assert every deployment's spec must have replicas between 1 and 4
		replicas!: int & >0 & <5

		selector!: matchLabels!: #Labels
		template!: {
			metadata!: labels!: #Labels
			spec: _
		}
		...
	}
}

deployments: {
	[string]: #Deployment
}

validated: deployment.#ValidateSelectorsOverMap & {map: deployments}
