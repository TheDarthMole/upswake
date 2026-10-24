ARG BUILDPLATFORM

FROM --platform=${BUILDPLATFORM} golang:1.27.3-alpine@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS build

WORKDIR "/build/"

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# To improve layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG GIT_DESCRIBE

RUN --mount=type=cache,target=/root/.cache/go-build,id=${TARGETOS}${TARGETARCH}${TARGETVARIANT} \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -tags "timetzdata" -trimpath -ldflags="-w -s -X 'main.Version=${GIT_DESCRIBE}'" \
    -o /opt/upswake/upswake ./cmd/upswake

FROM scratch AS minimal
COPY --from=build /opt/upswake/upswake /bin/
HEALTHCHECK --timeout=10s CMD ["/bin/upswake", "serve", "healthcheck"]
ENTRYPOINT ["/bin/upswake"]
