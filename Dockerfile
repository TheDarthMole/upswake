FROM golang:1.21.1-alpine as buld
COPY . /srv
RUN go build -o /srv/bin/UPSMon /srv/cmd/main.go

FROM scratch
COPY --from=build /srv/bin/UPSMon /srv/bin/UPSMon
WORKDIR /srv
ENTRYPOINT ["/srv/bin/UPSMon"]