FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY internal internal
COPY pkg pkg
COPY cmd cmd

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o kjournal cmd/*

FROM alpine:3.16 as kjournal-cli
WORKDIR /
COPY --from=builder /workspace/kjournal /usr/bin/
USER 65532:65532

ENTRYPOINT ["/usr/bin/kjournal"]
