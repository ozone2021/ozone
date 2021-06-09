FROM golang:1.16.4-alpine3.13 AS build-env

WORKDIR /go/src/app
COPY . .

RUN cd ozone/cmd/daemon && go build -o /go/src/app/ozone_daemon

# final stage
FROM ozone-daemon-base:latest
WORKDIR /app
COPY --from=build-env /go/src/app/ozone_daemon /app/

ENTRYPOINT ./ozone_daemon

EXPOSE 8000/tcp