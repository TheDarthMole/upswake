# Default shell for script-like behavior
set shell := ["bash", "-cu"]
set unstable

export GIT_DESCRIBE := `git describe --tags --always --long`
BUILD_ARGS := '-tags "timetzdata" -trimpath -ldflags="-w -s -X main.Version=${GIT_DESCRIBE}"'

# Display this help message
help:
    just -l

# Run all Go tests locally
test-local: generate-certs start-nut-server test && stop-nut-server

# Run all Go tests, assuming the NUT server is already running and certs are generated
test:
    go clean -testcache
    go test -coverpkg=$(go list ./... | grep -v 'internal/mocks' | tr '\n' ',') -coverprofile=coverage.txt -race -v ./...

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

_check-container-tool:
    {{if container-tool == "" { error("Neither podman nor docker was found in PATH. Please install one or set the CONTAINER_TOOL environment variable")} else { "" } }}

# Install development dependencies
install-deps: && _check-container-tool
    go install github.com/swaggo/swag/cmd/swag@latest
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Run upswake with arguments
run *args:
    go run {{BUILD_ARGS}} ./cmd/upswake {{args}}

# Build upswake
build:
    go build {{BUILD_ARGS}} -o ./upswake ./cmd/upswake

# Build the thedarthmole/upswake:local container
build-container: _check-container-tool
    echo ${GIT_DESCRIBE}
    {{container-tool}} build -t thedarthmole/upswake:local -f ./Containerfile . --build-arg GIT_DESCRIBE=${GIT_DESCRIBE}

# Builds and runs the upswake container
run-container: build-container
    CONTAINER_TAG=local {{container-tool}} compose up --force-recreate

# Runs a NUT server in a container for testing
start-nut-server: _check-container-tool
    {{container-tool}} compose -f hack/nut/compose.yaml up --force-recreate --remove-orphans --detach

# Stops the NUT server container
stop-nut-server: _check-container-tool
    {{container-tool}} compose -f hack/nut/compose.yaml down

generate-certs:
    mkdir -p certs
    openssl req -nodes -new -x509 -keyout certs/rsa.key -out certs/rsa.cert \
        -subj "/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
    openssl ecparam -out certs/ecc.key -name prime256v1 -genkey
    openssl req -new -x509 -key certs/ecc.key -out certs/ecc.cert \
        -subj "/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
