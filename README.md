# ozone
Environment variable management.


1. `mkdir /tmp/ozone`
2. run daemon
3. use ozone
## ozone-daemon docker 
`docker run --user root --rm -v /var/run/docker.sock:/var/run/docker.sock -d -t -v /tmp/ozone:/tmp/ozone -p 8000:8000 --name ozone-daemon -listen=:8000 ozone-daemon`   

`docker exec -it (docker run -v /var/run/docker.sock:/var/run/docker.sock --network host -d ozone-daemon) /bin/sh`   

`docker ps | grep ozone | awk '{print $1}' | xargs -I {} docker kill {}`   


`ping host.docker.internal` to find host ip   
// TODO command to add registry.local to host ip

curl https://registry.local/v2/_catalog -k

go get -u github.com/ozone2021/ozone/ozone

`Build debug ozone container`

docker build . -t ozone-daemon-base --progress plain -f Dockerfile.base;  

docker rm -f ozone-daemon; docker build . -t ozone-daemon --progress plain && docker exec -it (docker run --user root --restart=always -v /var/run/docker.sock:/var/run/docker.sock -d -t -v /tmp/ozone:/tmp/ozone -p 8000:8000 --name ozone-daemon -listen=:8000 ozone-daemon) /bin/sh

ga -A; gc "Latest."; git tag -d 1.2 && git push --delete origin 1.2; git tag 1.2; git push --tags
go install  github.com/ozone2021/ozone/ozone@1.2