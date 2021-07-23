## Ozone

Ozone is a build pipeline management system, predominantly aimed at developers that work with docker / k8s.   

Ozone can:
 - manage environment variables at the top level, using composable blocks of dynamic environments that can be merged together.
 - build + run services (dockerfiles) as if they were on your localhost
 - ignore specified services, to use your IDE and debugger to interface with the services instead.
 - cache (ignore) directories that haven't changed since the last build, for faster build times.


### *Installation*

You will need the ozone daemon running, and also the ozone client.



#### *Daemon*
You can either:
##### Run from docker hub...
`docker run --user root --restart=always -v /var/run/docker.sock:/var/run/docker.sock -d -t -v /tmp/ozone:/tmp/ozone -p 8000:8000 --name ozone-daemon -listen=:8000 ozone2021/ozone-daemon:latest`
 TODO does ozone folder need created in TMP?
##### ...or build locally
`docker rm -f ozone-daemon; docker build . -t ozone-daemon --progress plain; docker run --user root --restart=always -v /var/run/docker.sock:/var/run/docker.sock -d -t -v /tmp/ozone:/tmp/ozone -p 8000:8000 --name ozone-daemon -listen=:8000 ozone-daemon`

The daemon runs on port 8000. This isn't open to configuration, as of yet.

The daemon is used to keep track of cache hashs to tell whether or not any runnable needs to run again.

`docker logs ozone-daemon -f` is a useful way to keep track on what's going on.

#### *Client*

Easy:   
`go get -u github.com/ozone2021/ozone/cmd/main`


# Getting started

Ozone is for making environment management easier. The best way to learn is by doing, so there is a test repo which we will also use for the tutorial.
```markdown
git clone git@github.com:ozone2021/ozone-test.git --recurse-submodules

cd ozone-test

# For linux
git checkout -b tutorial/1-simple-runnable-ubuntu; git submodule update

# For mac
git checkout -b tutorial/1-simple-runnable-macos; git submodule update
```
Docker's host network mode works on ubuntu but not on mac, so this is why the `NETWORK` is set to the project name in 
the mac version.

## 1. A simple runnable

The Ozonefile is supposed to be a full specification of how your project runs, including all of the environments.

Have you ever joined a team, tried to run the project locally, and it fails with some unhelpful message?

This is the bane of interpretated language developer's lives, and why I prefer golang to nodejs/ruby/python. 

More times than enough, this error is because of environment variables. Which don't exist on your fresh machine, due to 
the .env file the other developer is using not being checked in (.gitignore). Using ENV variables in code without checking their 
existence at program startup only compounds this already unadulterated chaos. Developers need to then read through code to find which 
ENV var they need to make it work. Future versions of Ozone will provide a startup bottleneck that checks for all the defined required variables 
before a program runs.

An Ozonefile decouples the environment providing mechanism (k8s secrets file / .env files) from the build command. In a 
similar fashion to GNU Make, the best practice for Ozonefiles is to create an "all" runnable. Runnables are similar to 
targets in Make.

##### Runnables have 3 types, and build in the following order:
1. builds
2. deploys
3. tests

The only thing that distinguishes the 3 types is their order of execution.

We can run:

`ozone r d-micro-a b-micro-a`

And the b-micro-a runnable will run first, as it is a build.

Run:

`ozone r all`

If you look at the Ozonefile, you will see that `all` will run the same runnables as the previous command.

The caching mechanism saved the combined hash of the Ozonefile and the given directory.

The "dir" and "service" attributes must be set for this to work.

Caching is only used for build stages, this is because the caching of environment variable files hasn't yet been 
implemented, so whenever variables are changed, the system cannot tell, so deploys are run each time to remedy this.

#### Status
`ozone s`
Shows the port service and port

#### Logs

Plans to multiplex logs in future, for now you can use the docker logs of the container. The container is named after 
service.

`docker logs micro-a -f`

## 2. Different contexts

```
# For linux
git checkout -b tutorial/2-different-contexts-ubuntu; git submodule update

# For mac
git checkout -b tutorial/2-different-contexts-macos; git submodule update~~~~
```

A context is similar to what normally developers would call "environments", however in Ozone, environments refer to 
blocks of environment variables. To distinguish between them, we have the term "context".

To view the current context,

`ozone c`

This will default to the defined default in the Ozonefile, if there is no context saved in memory in the daemon.

To set a different context:

`ozone c local-k8s`

This relies on you having a k8s cluster setup, with a docker registry running inside it, with the ingress address 
`registry.local` pointing at it.

Running `ozone r all` now, you will see the different context flash up.

"context_steps" and "context_envs" allow us to specify context specific steps and environment blocks.

Take the following example:
```
context_envs:
- context: local-k8s|ns-k8s  # helm-base is applied to local-k8s|ns-k8s
  with_env:
    - helm-base
- context: "{{CONTEXT}}"  # then if an environment exists with the same name as the context, it is applied
  with_env:
    - "{{CONTEXT}}"
```

It is a good practice to name environments after contexts, as this provides a building block from which to add more 
environments (which tend to be more step specific). 


What you think the following block does?:

```
context_steps:
- context: "{{CONTEXT}}"
  steps:
    - type: builtin
      name: buildDockerImage
- context: local-k8s|ns-k8s
  steps:
    - type: builtin
      name: pushDockerImage
```

For all contexts in the context list, (local, local-k8s and ns-k8s) it builds the docker image, but only for the 
kubernetes related contexts does it push it to the registry in the cluster.


#### Commandline arguments

Currently, ozone supports passing the context in via the command line, as well as specifying "headless" mode.

Headless or "detached" mode means you can run ozone without the daemon, which means you can't make use of caching. 
This might be desired behaviour if you are using ozone in the context of a build server that doesn't allow for the 
docker daemon to be installed.

#### Build deploy test

If you have [newman](https://www.npmjs.com/package/newman) installed, you can run the following command:

`ozone r -c ns-k8s -d all-test`

This sets the context to ns-k8s (namespaced kubernetes), in detached mode, with the runnable all-test. If you inspect the runnable, I like to 
put the aggregate builds "all" and "all-test" in the deploys section.

The last part of this, is that it runs automated tests against the health check endpoints.

A simple but effective test that should be the basis of any CI/CD pipeline. This will test for any runtime failures with
 interpretated languages like ruby. 
 
The output from the helm command is visible. You can see that the namespace is a hash. This is generated from the branch 
name in the parent project.

The container image tag for the microservices is generated from the commit hash of each respective git submodule.