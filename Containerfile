FROM --platform=${BUILDPLATFORM} golang:1.25.1-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS build

WORKDIR "/build/"

# To improve layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG GIT_DESCRIBE
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG COMMIT_SHA
ARG BUILD_DATE

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -tags "timetzdata" -trimpath -ldflags="-w -s -X 'main.Version=${GIT_DESCRIBE}' -X 'main.Commit=${COMMIT_SHA}' -X 'main.Date=$(date '+%Y-%m-%d %H:%M:%S %z')'" \
    -o /opt/upswake/UPSWake ./cmd/upswake

FROM scratch AS minimal
COPY --from=build /opt/upswake /
ENTRYPOINT ["/UPSWake"]