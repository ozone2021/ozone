# ozone
Environment variable management.

## ozone-daemon docker 
`docker run --user root --rm -v /var/run/docker.sock:/var/run/docker.sock -d -t -p 8000:8000 --name ozone-daemon -listen=:8000 ozone-daemon`   

`docker exec -it (docker run -v /var/run/docker.sock:/var/run/docker.sock --network host -d ozone-daemon) /bin/sh`   

`docker ps | grep ozone | awk '{print $1}' | xargs -I {} docker kill {}`   

`docker run -p 8080:8080 hashicorp/http-echo -listen=:8080 -text="hello world"`   


`ping host.docker.internal` to find host ip   
// TODO command to add registry.local to host ip

curl https://registry.local/v2/_catalog -k