FROM --platform=${BUILDPLATFORM} golang:1.23.2-alpine AS build

WORKDIR "/srv/"

# To improve layer caching
COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY [".", "/srv/"]

ARG GIT_DESCRIBE
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -tags "timetzdata" -trimpath -ldflags="-w -s -X main.Version=${GIT_DESCRIBE} -buildid=" \
    -o ./UPSWake ./cmd/upswake

FROM scratch AS minimal
COPY --from=build /srv/UPSWake /UPSWake
COPY --from=build /srv/rules/ /rules/
ENTRYPOINT ["/UPSWake"]