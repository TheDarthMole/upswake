ARG BUILDPLATFORM

FROM --platform=${BUILDPLATFORM} golang:1.26.4-alpine@sha256:f1ddd9fe14fffc091dd98cb4bfa999f32c5fc77d2f2305ea9f0e2595c5437c14 AS build

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
