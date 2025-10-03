package deployment

#Labels: [string]: string

_deployment: {
	metadata!: {
		name!: string
		...
	}
	spec!: {
		selector!: matchLabels!: #Labels
		template!: {
			metadata!: labels!: #Labels
			spec: _
		}
		...
	}
	...
}

#ValidateSelectors: {
	d: _ & _deployment

	selectorEqualLabels: close({
		for k, v in d.spec.selector.matchLabels {"\(k)": v}
	})
	selectorEqualLabels: close({
		for k, v in d.spec.template.metadata.labels {"\(k)": v}
	})
}

#ValidateSelectorsOverMap: {
	map: [string]: _deployment
	validated: {
		for i, deploy in map {
			"deployment<\(deploy.metadata.name)>": #ValidateSelectors & {d: deploy}
		}
	}
}
