FROM --platform=${BUILDPLATFORM} golang:1.24.0-alpine3.20@sha256:79f7ffeff943577c82791c9e125ab806f133ae3d8af8ad8916704922ddcc9dd8 AS build

WORKDIR "/build/"

# To improve layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG GIT_DESCRIBE
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -tags "timetzdata" -trimpath -ldflags="-w -s -X main.Version=${GIT_DESCRIBE} -buildid=" \
    -o /opt/upswake/UPSWake ./cmd/upswake

FROM scratch AS minimal
COPY --from=build /opt/upswake /
ENTRYPOINT ["/UPSWake"]