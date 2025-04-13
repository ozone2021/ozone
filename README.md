# Ozone

Environment variable management and build system framework.

## Installation

```
git clone git@github.com:ozone2021/ozone.git
./script/build_run_daemon.sh
go install  github.com/ozone2021/ozone/ozone@latest
```

## What does ozone do?

Most developers are familiar with GNU Make. How would you describe what GNU Make does?

People generally use Makefiles to build C or C++ projects, but it can be used for any compiled language. Some people use 
Makefiles as a wrapper around terraform/infrastructure. Why? Because there is inherent value in having a familiar interface to interact with your codebase. Having a folder full of scripts 
isn't very useful, because sometimes scripts aren't supposed to be called by themselves, they are supposed to be called 
by other scripts and it is assumed at the point in time that the script is called, that certain environment variables have 
already been set by the calling script.

### The general use case

Ozone provides a common interface to interact with your codebase. It can be used to build and deploy containers, but also 
provides a way of providing one click installations that take care of your infrastructure dependencies for when you're
building locally, for example on kubernetes.

It comes with a built-in caching mechanism that has almost instantaneous cache hits. Many other frameworks such as skaffold 
rely on the caching mechanism of whichever container builder you're using, which is often slow, because for eg docker 
relies on the entire build context being copied to the docker daemon. If your root project folder is large, you're multiplying 
that by the number of services and then having to hold that in memory, just to go through the docker build cache and be told 
nothing has changed. This alone can add a minute or 2 to your build time.

### Decoupling from CI/CD

Too often, people write their CI/CD pipelines in a way that the build is tightly coupled to the CI/CD provider. Their scripts
cannot be reused locally, and in the worst case scenario, developers have to push their code to a branch and wait for the CI/CD pipeline to run 
in order for the containers to be built, so that they can manually copy the built image tag into their kubernetes manifests. This is 
as far away from automated deployments as you can get. Automated end to end testing is impossible, and you have to rely
on human developers to run the end to end testing each time they make a change, which results in human error.

Ozone encourages decoupling of the scripts from the CI/CD 

## Environment variable management

A lot of people think of environments as dev/staging/production. This completely ignores the most vital aspect of development,
which is the developer's local machine, and secondarily, that many developers can't fully test their work locally, so the
need for pull request generated sandbox (preview) environments becomes necessary, so that their code can be tested thoroughly before being merged 
into the development branch.

Many lists of environment variables are common between different contexts.

For example:

```
TEST_ENDPOINTS_ENABLED: "true"
CONFIGMAP_FILENAMES: "dev-configmap.yaml"
CONTAINERISED_DB: "true"
```

The above environment variables are common between local/sandbox/dev.
How can we share environment variables without massive repetition?

Yaml anchors can be used depending on the framework but often break IDE yaml linting or golang parsing. They also don't support templating.

```yaml
  build_vars:
    NON_LIVE_CONTEXTS: 'local|pull-requests|dev'  <--- regex pattern
    
  environments:
    - name: non-live-contexts
      with_vars:
        TEST_ENDPOINTS_ENABLED: "true"
        CONFIGMAP_FILENAMES: "dev-configmap.yaml"
        CONTAINERISED_DB: "true"

  deploys:  
    - name: deploy
      context_envs:
        - context: "{{NON_LIVE_CONTEXTS}}}}"  <--- pattern matching
          with_env:
            - non-live-contexts
```

We can list the variables that are common between different contexts in a single place (underneath environments), and then use pattern matching on the 
context_envs to include the common variables.

## Documentation Index

- [Config](./docs/config.md)
- [Built in environments](./docs/built_in_environments.md)
- [Commands](./docs/commands.md)
