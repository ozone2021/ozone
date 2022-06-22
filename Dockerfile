FROM golang:1.18.3-alpine AS build-env

WORKDIR /go/src/app
COPY . .

RUN cd ozone/cmd/daemon && go mod vendor && go build -o /go/src/app/ozone_daemon

# final stage
FROM ozone-daemon-base:latest
WORKDIR /app
COPY --from=build-env /go/src/app/ozone_daemon /app/

ENTRYPOINT ./ozone_daemon

EXPOSE 8000/tcp