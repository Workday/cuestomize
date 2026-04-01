FROM golang:1.26 AS builder

ARG GOOS=linux
ARG LDFLAGS
ARG GOARCH=amd64

WORKDIR /workspace

ADD go.mod go.mod
ADD go.sum go.sum

RUN go mod download

ADD internal/ internal/
ADD api/ api/
ADD pkg/ pkg/
ADD main.go main.go

RUN CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags "$LDFLAGS" -o cuestomize main.go


FROM gcr.io/distroless/static:latest

COPY --from=builder /workspace/cuestomize /usr/local/bin/cuestomize

ENTRYPOINT ["/usr/local/bin/cuestomize"]
