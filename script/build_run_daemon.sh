#!/bin/bash


docker build . -t ozone-daemon-base --progress plain -f Dockerfile.base;

docker rm -f ozone-daemon; docker build . -t ozone-daemon --progress plain && docker exec -it $(docker run --user root --restart=always -v /var/run/docker.sock:/var/run/docker.sock -d -t -v /tmp/ozone:/tmp/ozone -p 8000:8000 --name ozone-daemon -listen=:8000 ozone-daemon) /bin/sh
