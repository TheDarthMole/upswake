ARG BUILDPLATFORM

FROM --platform=${BUILDPLATFORM} golang:1.25.6-alpine@sha256:d9b2e14101f27ec8d09674cd01186798d227bb0daec90e032aeb1cd22ac0f029 AS build

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
    -o /opt/upswake/upswake ./cmd/upswake

FROM scratch AS minimal
COPY --from=build /opt/upswake/upswake /bin/
HEALTHCHECK --timeout=10s CMD ["/bin/upswake", "serve", "healthcheck"]
ENTRYPOINT ["/bin/upswake"]
