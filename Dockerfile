FROM --platform=${BUILDPLATFORM} golang:1.23.5-alpine3.20@sha256:def59a601e724ddac5139d447e8e9f7d0aeec25db287a9ee1615134bcda266e2 AS build

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