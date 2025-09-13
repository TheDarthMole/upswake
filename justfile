# Default shell for script-like behavior
set shell := ["bash", "-cu"]
set unstable

# Display this help message
help:
    just -l

# Run all Go tests
test: start-nut-server && stop-nut-server
    go clean -testcache
    go test ./...

# Runs all linters
lint: swagger
    swag fmt .
    golangci-lint fmt
    golangci-lint run
    go mod tidy

# Generate Swagger documentation
swagger:
    swag init --parseDependency \
        --parseInternal \
        --parseDepth 1 \
        -d "./internal/api/handlers" \
        -g "./root.go" \
        -o "./internal/api/docs"

# Define the container tool with auto-detection, or allow override via CONTAINER_TOOL
container-tool := if env("CONTAINER_TOOL", "") != "" { env("CONTAINER_TOOL")
} else if which("podman") != "" { "podman"
} else if which("docker") != "" { "docker"
} else {""}

# Install development dependencies
install-deps:
    go install github.com/swaggo/swag/cmd/swag@latest
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    {{if container-tool == "" { error("Neither podman nor docker was found in PATH. Please install one or set the CONTAINER_TOOL environment variable")} else { "" } }}

# Run upswake with arguments
run *args:
    go run ./cmd/upswake {{args}}

# Build upswake
build:
    go build ./cmd/upswake -o ./upswake

# Build the thedarthmole/upswake:local container
build-container:
    {{if container-tool == "" { error("Neither podman nor docker was found in PATH. Please install one or set the CONTAINER_TOOL environment variable")} else { "" } }}
    {{container-tool}} build -t thedarthmole/upswake:local -f ./Containerfile .

# Builds and runs the upswake container
run-container: build-container
    CONTAINER_TAG=local {{container-tool}} compose up --force-recreate

# Runs a NUT server in a container for testing
start-nut-server:
    {{if container-tool == "" { error("Neither podman nor docker was found in PATH. Please install one or set the CONTAINER_TOOL environment variable")} else { "" } }}
    {{container-tool}} compose -f hack/nut/compose.yaml up --force-recreate --remove-orphans --detach

# Stops the NUT server container
stop-nut-server:
    {{if container-tool == "" { error("Neither podman nor docker was found in PATH. Please install one or set the CONTAINER_TOOL environment variable")} else { "" } }}
    {{container-tool}} compose -f hack/nut/compose.yaml down

generate-cert:
    mkdir -p certs
    openssl req  -nodes -new -x509 -keyout certs/server.key -out certs/server.cert \
        -subj "/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
