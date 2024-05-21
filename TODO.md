Check for optional args in steps, and flag if mispelt etc.
eg, WORKING_DIR in bashScript

Overriding context env var from with_var of calling runnable doesn't override.

  - name: deploy-webserver
    service: webserver
    dir: "{{SUBSCRIPTION_DIR}}"
    depends_on:
      - name: deploy-microservice
        with_vars:
          VALUES_FILE: "infrastructure/helm/generic_values.yaml"
          CHART_DIR: "{{HELM_CHARTS_DIR}}/override" <--- this doesn't override





Version check on Ozonefile.

ContextStep step -> should probably force name being set to give better error messages.

document:
Conditional script should return 3 for exit that is a real error. All others should be ignored.

# Bugs

#### If LOCAL_CONTEXT isn't set in build_vars, silent failure.
```
context_steps:
  - context: "{{LOCAL_CONTEXT}}"
```

### if no default env that matches context name, fails

## Script with_vars

with_vars aren't rendered.

#### Breaks if build_vars empty.

#### Caching doesn't take into account external vars changing, nor any source files of called runnables.

Hashing the entire runspec would fix external vars.

Source files need merged all the way down to the bottom of the call stack.

The resultant hash should be added to the root of the call stack.

# Refactoring

fetchEnv should use a stack instead of recursion.

refactor cmd/run.go

## Custom yaml unmarshall should unmarshal to Variables and not just string/[]string

# Features

## Env name prefix

Special Env var PREFIX would prepend something to all env vars, for easier debugging / grepping (isolate from system env vars)

# Error handling

## Check for duplicate runnable names

## Passing empty string to build_vars messes up default

CONFIGMAP_FILENAME: '{{CONFIGMAP_FILENAME | default_if_none:"dev-env-configmap.yaml"}}'

^if CONFIGMAP_FILENAME is passed from eg github actions as empty, it messes up.

#### helm deployable
argsVar.GetStringValue() doesn't error when it's a slice

#### ozone -d -c  run deploy-all
no context passed but it doesn't error

#### context_condition should be context_conditionals

And throw error if incorrect

#### Context not found throws segfault

#### env/env/git_branch_ticket_number
Double env at the start typo doesn't throw error for non existent env.

#### this works
build_vars:
  services:
    - NODEJS
    - ML-AI

#### this doesn't. Needs hyphens. No decent error thrown.
build_vars:
  services:
    NODEJS
  [Ozonefile](..%2Fdeleteme%2FOzonefile)  ML-AI


# Document

#### List addition works now

#### sourcefileprepend
only prepends to the source files for the callstack, not for the ones in the buildscope, it stays relative.

# plugin capability
https://github.com/hashicorp/go-plugin

# Caching

Changing the ozonefile triggers rebuild of all services. Use the scopes as part of the hash, and the serialization of the
runnable itself. This meanings

# Testing

#### Rendering of build vars needs tested.

#### Test context conditional.
What about scripts with args?


## Github action

### To install

Deffo.

### To run?

Probably

# Change license so that any paid services must pay to use 

# High priority

Use cancellable context for all commands so it finishes when ctrl-c is pressed.

# LogApp

Find way to update runResult as soon as log app connects.

Delete log files on start of new run.

Page down as new logs come in.


# builtin error

```
  - name: test-ozone
    dir: "{{API_DIR}}"
    steps:
      - type: builtin
        name: runScript
        with_vars:
          SCRIPT: script/test_ozone.sh
```
runScript doesn't exist, but exits with success.

# Log fix

Reusing the same file causes mad issues.

Create a folder for each run, and delete the folder when the run is finished.

# Directory bug

The /tmp/ozone/<dir> uses the base64 encoding of the ozone working directory. If this is too long, the unix pipe fails 
to open with:

```
2024/05/13 11:29:42 failed to listen: listen unix /tmp/ozone/L1VzZXJzdZVyLXBvYw==/socks/log-b61cbd.sock: bind: invalid argument
```

