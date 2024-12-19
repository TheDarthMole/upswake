FROM --platform=${BUILDPLATFORM} golang:1.23.4-alpine3.20@sha256:9a31ef0803e6afdf564edc8ba4b4e17caed22a0b1ecd2c55e3c8fdd8d8f68f98 AS build

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