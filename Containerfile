from golang:1.25 as builder

workdir /workspace

add go.mod go.mod
add go.sum go.sum

run go mod download

add internal/ internal/
add api/ api/
add pkg/ pkg/
add main.go main.go

add semver semver

run CGO_ENABLED=0 go build -o cuestomize main.go

from gcr.io/distroless/static:latest 

copy --from=builder /workspace/cuestomize /usr/local/bin/cuestomize

entrypoint ["/usr/local/bin/cuestomize"]
