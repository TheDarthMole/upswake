FROM golang:1.21.1-alpine as build
COPY . /srv
WORKDIR /srv
RUN go get
RUN go build -o ./UPSWake ./

FROM scratch
COPY --from=build /srv/UPSWake /UPSWake
WORKDIR /srv
ENTRYPOINT ["/UPSWake"]