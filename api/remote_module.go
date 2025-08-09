package api

import "sigs.k8s.io/kustomize/api/types"

type RemoteModule struct {
	Registry string `yaml:"registry" json:"registry"`
	Repo     string `yaml:"repo" json:"repo"`
	Tag      string `yaml:"tag" json:"tag"`

	Auth      *types.Selector `yaml:"auth,omitempty" json:"auth,omitempty"`
	PlainHTTP bool            `yaml:"plainHTTP,omitempty" json:"plainHTTP,omitempty"`
}
