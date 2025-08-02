FROM golang:1.24 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -o cuestomize main.go


FROM gcr.io/distroless/static:latest

COPY --from=builder /workspace/cuestomize /usr/local/bin/cuestomize

ENTRYPOINT ["/usr/local/bin/cuestomize"]
