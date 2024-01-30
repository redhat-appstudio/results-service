# Build the manager binary
FROM registry.access.redhat.com/ubi9/go-toolset:1.20.10-11@sha256:c7c3cf77133ff99b32e04ba1b6f238d15ed97403dae1a24e5b3f6ad82087f07e as builder

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY vendor/ vendor/
COPY pkg/ pkg/
COPY cmd/ cmd/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o results-service ./cmd/service

# Use ubi-minimal as minimal base image to package the manager binary
# Refer to https://catalog.redhat.com/software/containers/ubi8/ubi-minimal/5c359a62bed8bd75a2c3fba8 for more details
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.3-1552@sha256:582e18f13291d7c686ec4e6e92d20b24c62ae0fc72767c46f30a69b1a6198055
COPY --from=builder /opt/app-root/src/results-service /
USER 65532:65532
EXPOSE 8443
ENTRYPOINT ["/results-service"]
