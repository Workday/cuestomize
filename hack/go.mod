module github.com/Workday/cuestomize/hack

go 1.25.0

require (
	github.com/Workday/cuestomize v0.0.0
	github.com/rs/zerolog v1.34.0
	oras.land/oras-go/v2 v2.6.0
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
)

replace dagger/cuestomize => ../.dagger

replace github.com/Workday/cuestomize => ../.
