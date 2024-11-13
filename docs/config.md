# Ozonefile (Config)

The Ozonefile is parsed into a ConfigStruct in `ozone-lib/config/config.go`.

```go
type OzoneConfig struct {
    Headless      bool           `yaml:"headless"`
    ProjectName   string         `yaml:"project"`
    ContextInfo   ContextInfo    `yaml:"context"`
    BuildVars     *VariableMap   `yaml:"build_vars"`
    Environments  []*Environment `yaml:"environments"`
    PreUtilities  []*Runnable    `yaml:"pre_utilities"`
    Builds        []*Runnable    `yaml:"builds"`
    Deploys       []*Runnable    `yaml:"deploys"`
    Tests         []*Runnable    `yaml:"tests"`
    Pipelines     []*Runnable    `yaml:"pipelines"`
    PostUtilities []*Runnable    `yaml:"post_utilities"`
}
```

Every named runnable can be called independently, or they can be passed into `ozone run` as a list of runnables to run in sequence.

The ordering of the runnables in the CLI params is mostly unimportant. 

They always run in the order that they are shown in the struct:

- PreUtilities
- Builds
- Deploys
- Tests
- Pipelines
- PostUtilities

Each runnable type can only call other runnables of the same type.

Pipeline runnables are a special case, they can call any other type of runnable. This provides readable in-code documentation 
form how to run each pipeline, rather than the alternative of having to document separately which bash script to use where
and how.

### Headless mode (Daemonless)

Headless mode is a mode where the ozone daemon is not started. The docker daemon is only meant for local development.

The daemon keeps track of the caching hashes. See more in [Cache](#cache).

### Contexts

A `context` in Ozone is similar to what people would generally call an environment.

An `environment` in Ozone is a key:value list of environment variables. When a list contains another list, the inner list is merged with the outer list.

Running ozone as a different context, allows you to develop your CI/CD pipeline from your local machine, this is a huge time
saving measure because the startup time of VMs on github actions/circleci is slow, and you can't easily debug things.

Ozone allows you to develop each invididual step of your pipeline from your local machine, rather than having to wait for the
entire pipeline to run while adding new steps.

### BuildVars (Build Variables)

VariableMap is map of strings to either string or []string. List variables get rendered as a semi-colon separated list.

This section is used to define the top level variables. Often when people use .env files, they don't have an easy way to
share the .env file as it doesn't get checked into source. Too often, there's a lack of documentation of what variables are 
needed for the software at runtime.

This same problem presents itself at build time, by defining all the variables at the top, it provides a self documenting
description of what is needed.

Variables here are rendered using [pongo2 templating library](https://github.com/flosch/pongo2).

The templating library allows the setting of defaults for if an environment variable isn't set. Normally you set the
defaults to be for local development, then override them in the CI/CD, as CI/CD is set up once, whereas local development
has to be set up for each developer, so it makes more sense to make the setup as simple as possible for local.

#### Variable ordinality

Normally in programming languages, the scope of variables is like this:

```
x = 0
if true {
    x = 1
    // x will be 1
}
```

This is not the case for ozone. Ordinality dictates which variables can overwrite which.

BuildVariables have the highest Ordinal, which is integer.MAX_INT.

The initial root callstack runnables have ordinal 1.

### Environments

```go
type Environment struct {
    Name     string       `yaml:"name"`
    For      *For         `yaml:"for,omitempty"`
    WithVars *VariableMap `yaml:"with_vars"`
    Includes []*Include   `yaml:"include"`
}

type Include struct {
    Name      string       `yaml:"name"`
    InputVars *VariableMap `yaml:"input_vars"`
    Type      string       `yaml:"type"`
}
```

Environments are a way to group variables together. They can be included in other environments.

An include is a way to include another environment, but with the ability to override the variables.

For example:

```yaml
- name: base_secret_file
  include:
    - name: env/from_k8s_secret_file
      type: builtin
      input_vars:
        SECRET_FILE: ./secret.yml

- name: local
  include:
    - name: base_secret_file
  with_vars:
    KEY: "value"
    LIST_EXAMPLE:
      - item1
      - item2
```

The local environment includes base_secret_file, which is a custom environment.

You can also make use of [built in environments](./built_in_environments.md). These are environments that are included in the ozone library.

#### Loops / Variable variables

You can loop over a list of items to create environments.

```yaml
build_vars:
  services:
   - service1
   - service2
   - healthchecker
environments:
  - name: all-docker-tags-latest
    for: service:services
    with_vars:
      "{{service | upper}}_FULL_TAG": "{{DOCKER_REGISTRY}}/{{service}}:latest"
```

Running ozone plan
```bash
$ ozone plan
```
Will show this:
```yaml
...
scope: |
  - name: SERVICE1_FULL_TAG
    value: registry.local/service1:latest
    ordinal: 1
  - name: SERVICE2_FULL_TAG
    value: registry.local/service2:latest
    ordinal: 1
  - name: HEALTHCHECKER_FULL_TAG
    value: registry.local/healthchecker:latest
    ordinal: 1
  
...
```

N.B currently pongo2 doesn't support a filter to convert to pascal case, so you can't use hyphens in the service names.


### Runnables

Run in this order: 

```
PreUtilities
Builds
Deploys
Tests
Pipelines
PostUtilities
```

```go
type Runnable struct {
	Parallel            *bool                 `yaml:"parallel,omitempty"`
	Name                string                `yaml:"name"`
	DropContextEnv      bool                  `yaml:"drop_context_env"`
	Cache               bool                  `yaml:"cache"`
	Service             string                `yaml:"service"`
	Dir                 string                `yaml:"dir"`
	SourceFiles         []string              `yaml:"source_files"`
	SourceFilesPrepend  string                `yaml:"source_files_prepend"`
	WithVars            *VariableMap          `yaml:"with_vars"`
	ContextEnv          []*ContextEnv         `yaml:"context_envs"`
	ContextConditionals []*ContextConditional `yaml:"context_conditionals"`
	Depends             []*Step               `yaml:"depends_on"`
	ContextSteps        []*ContextStep        `yaml:"context_steps"`
	Steps               []*Step               `yaml:"steps"`
	Type                RunnableType
}
```

#### Parallel

Runs all the dependent runnables in parallel.

```yaml
  - name: build-all
    parallel: true
    depends_on:
      - name: b-micro-a
      - name: b-micro-b
      - name: b-health-checker
```

#### Name

The name of the runnable, this can be called from the cli using `ozone run` TODO link

#### DropContextEnv

By default, if there is an environment name that matches the context, it is automatically included.

Eg for context called local, it will include the environment called local.

DropContextEnv will drop the context environment if set to true.

#### Cache

This is for local only, you need to make use of conditionals to check for images in a remote docker registry
for CI/CD.

Uses a hash of the last update times of the source files to determine whether to rebuild or not.

If any of the last update times have changed since the last successful build, the cache is invalidated and the runnable
will run again.

The scope (variables etc from ozone at the point it is called) is also hashed as well, so if any
relevant variables change, the cache is invalidated.

#### Service (optional)

The name of the service the runnable relates to.

It's optional. It doesn't need set for runnables that don't relate to a specific service, like `build-all`.

For anything related to a specific service, you need to set it for some of the built in environments or runnables to work.

#### Dir

The relative directory of the services codebase.

#### SourceFiles

List of the source files that make up the runnable (for caching and also for the git based hash creation).

To list all files in a directory, you have to use double asterisk as wildcard.

```yaml
source_files:
  - "./directory/**"
```

#### SourceFilesPrepend

A path to prepend to the start of the source files, so that you can list them relative to that path.

#### WithVars

Variables to pass to the runnable steps.

#### ContextEnv

Which specific environments to include for this runnable, for a given context.


#### ContextConditionals

Only run the runnable in certain contexts if the condition is met.

This is useful for running eg a script when on CI/CD to check if the container exists 
in a remote docker registry.

when runs if script succeeds
```yaml
context_conditionals:
  - context: "{{NS_CONTEXTS}}"
    when_script:
      - script/true.sh
```

when_not runs if script fails
```yaml
context_conditionals:
  - context: "{{NS_CONTEXTS}}"
    when_not_script:
      - script/aws_ecr_has_container.sh
```

#### DependsOn

The runnables that this runnable depends on.

```yaml
- name: build-all
  depends_on:
    - name: b-micro-a
    - name: b-micro-b
    - name: b-health-checker
```

#### ContextSteps

Steps that are specific to a context.

```yaml
context_steps:
  - context: "{{LOCAL_CONTEXT}}"
    steps:
      - script: script/something_that_only_happens_for_local.sh
```

#### Steps

```go
type Step struct {
	Type        string            `yaml:"type"`
	Name        string            `yaml:"name"`
	WithVars    *VariableMap      `yaml:"with_vars"`
	VarOutputAs map[string]string `yaml:"var_output_as"`
}
```

type is always builtin for now, you can't call other runnables as steps.

Different options are available for different runnableTypes:

### Name

##### PreUtility | PostUtility
- bashScript

##### Builds:
- buildDockerImage
- bashScript
- pushDockerImage
- tagDockerImageAs

##### Deploys:
- helm
- bashScript

##### Tests:
- bashScript

#### How to use the builtins:

##### bashScript

Run a bash script.

Input vars:
- SCRIPT: The path to the script.
- WORKING_DIR (optional): The working directory to run the script in. Defaults to root directory.

##### buildDockerImage

Build a docker image.

Input vars:
- DOCKER_FULL_TAG: The full tag (registy/image:tag) for the docker image.
- DOCKER_BUILD_DIR (optional): The directory to build the docker image from. Defaults to root directory.
- DOCKER_BUILD_ARGS (optional): The build args to pass to the docker build command.
- DOCKERFILE (optional): The path to the dockerfile. Defaults to `Dockerfile`.

##### pushDockerImage

Push a docker image.

Input vars:
- DOCKER_FULL_TAG: The full tag (registy/image:tag) for the docker image.

##### tagDockerImageAs

Retag a docker image.

Input vars:
- SOURCE_TAG: The source tag (registy/image:tag) for the docker image.
- TARGET_TAG: The target tag (registy/image:tag) for the docker image.

### WithVars

Variables to pass to the step.

### VarOutputAs

Output variables to set from the output of the step.