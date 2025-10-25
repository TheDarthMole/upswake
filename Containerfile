ARG BUILDPLATFORM

FROM --platform=${BUILDPLATFORM} golang:1.25.3-alpine@sha256:aee43c3ccbf24fdffb7295693b6e33b21e01baec1b2a55acc351fde345e9ec34 AS build

WORKDIR "/build/"

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# To improve layer caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build,id=build-${TARGETOS}${TARGETARCH}${TARGETVARIANT} \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG GIT_DESCRIBE

RUN --mount=type=cache,target=/root/.cache/go-build,id=build-${TARGETOS}${TARGETARCH}${TARGETVARIANT} \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -tags "timetzdata" -trimpath -ldflags="-w -s -X 'main.Version=${GIT_DESCRIBE}'" \
    -o /opt/upswake/UPSWake ./cmd/upswake

FROM scratch AS minimal
COPY --from=build /opt/upswake /
ENTRYPOINT ["/UPSWake"]
