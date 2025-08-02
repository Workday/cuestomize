package main

apiVersion: "cuestomize.dev/v1alpha1"
kind:       "NginxDeployment"

input: {
	deploymentName!: =~"^nginx-deployment(-[a-z0-9]+)?$"
	namespace!:      string
	image!:          string
	replicas:        int | *1
}

includes: {} | null

LBL=labels: {
	app:       input.deploymentName
	component: "nginx"
}

outDeployment: {
	apiVersion: "apps/v1"
	kind:       "Deployment"
	metadata: {
		name:      input.deploymentName
		namespace: input.namespace
	}
	spec: {
		replicas: input.replicas
		selector: {
			matchLabels: LBL
		}
		template: {
			metadata: {
				labels: LBL
			}
			spec: {
				containers: [{
					name:  "nginx"
					image: input.image
				}]
			}
		}
	}
}

outputs: [
	outDeployment,
]
