module github.com/Workday/cuestomize/e2e

go 1.25.0

require (
	dagger/cuestomize v0.0.0
	github.com/Workday/cuestomize v0.0.0
	oras.land/oras-go/v2 v2.6.0
)

require (
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	golang.org/x/sync v0.16.0 // indirect
)

replace dagger/cuestomize => ../.dagger

replace github.com/Workday/cuestomize => ../.
