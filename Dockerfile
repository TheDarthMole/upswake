FROM golang:1.21.1-alpine as build
COPY . /srv
WORKDIR /srv
RUN go get
RUN go build -o ./UPSWake ./

FROM scratch
COPY --from=build /srv/UPSWake /UPSWake
COPY --from=build /srv/rules/ /srv/rules/
WORKDIR /srv
ENTRYPOINT ["/UPSWake"]