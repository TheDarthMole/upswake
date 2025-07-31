FROM --platform=${BUILDPLATFORM} golang:1.24.5-alpine AS build

WORKDIR "/build/"

# To improve layer caching
RUN apk add --no-cache upx
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG GIT_DESCRIBE
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -tags "timetzdata" -trimpath -ldflags="-w -s -X main.Version=${GIT_DESCRIBE}" \
    -o /opt/upswake/UPSWake ./cmd/upswake

RUN upx -7 --no-backup /opt/upswake/UPSWake

FROM scratch AS minimal
COPY --from=build /opt/upswake /
ENTRYPOINT ["/UPSWake"]