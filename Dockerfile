FROM golang:1.23.2-alpine AS build
COPY . /srv
WORKDIR /srv
RUN go get
RUN go test ./...
RUN CGO_ENABLED=0 go build -o ./UPSWake ./

FROM scratch AS minimal
COPY --from=build /srv/UPSWake /UPSWake
COPY --from=build /srv/rules/ /rules/
WORKDIR /
ENTRYPOINT ["/UPSWake"]