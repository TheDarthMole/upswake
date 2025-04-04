FROM --platform=${BUILDPLATFORM} golang:1.24.2-alpine3.20@sha256:00f149d5963f415a8a91943531b9092fde06b596b276281039604292d8b2b9c8 AS build

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