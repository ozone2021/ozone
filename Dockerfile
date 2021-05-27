FROM golang:1.16.4-alpine3.13 AS build-env

WORKDIR /go/src/app
COPY . .

RUN cd cmd && ./goModSync.sh
RUN cd cmd/daemon && go build -o /go/src/app/ozone_daemon

# final stage
FROM alpine:3.13
WORKDIR /app
COPY --from=build-env /go/src/app/ozone_daemon /app/

RUN apk update; apk add docker vim curl
ENTRYPOINT ./ozone_daemon

EXPOSE 8000/tcp