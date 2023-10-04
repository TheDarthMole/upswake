FROM golang:1.21.1-alpine as build
COPY . /srv
WORKDIR /srv
RUN mkdir -p ./bin
RUN go build -o ./bin/UPSWake ./

FROM scratch
COPY --from=build /srv/bin/UPSWake /srv/bin/
WORKDIR /srv
ENTRYPOINT ["/srv/bin/UPSWake"]