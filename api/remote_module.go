package api

type RemoteModule struct {
	Repo string `json:"repo"`
	Tag  string `json:"tag"`

	Auth *RemoteAuth `json:"auth,omitempty"`
}

type RemoteAuth struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	RefreshToken string `json:"refreshToken,omitempty"`
	AccessToken  string `json:"accessToken,omitempty"`
}
