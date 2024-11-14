# Built in environments

Example format:
```yaml
environments:
  - name: base_secret_file
    include:
      - name: env/from_k8s_secret_file
        type: builtin
        input_vars:
          SECRET_FILE: ./secret.yml
```

## env/from_k8s_secret_file

Takes the kubernetes secret file and loads it into the environment.

Required input_vars:
- SECRET_FILE: The path to the secret file.

## env/from_k8s_secret64

Takes base64 encoded kubernetes secret and loads it into the environment.

Required input_vars:
- SECRET_BASE64: The base64 encoded secret.

## env/from_env_file

Loads the environment variables from a env file into the environment.

Required input_vars:
- ENV_FILE: The path to the env file.

## env/git_log_hash

Uses the git command `git log -n 1 --pretty=format:%%H --  %s`

This takes the list of all files as an input, and gets the most recent hash of all the files.

This is for use in CI/CD to create a unique hash for the build. It will change whenever the source 
files change, meaning that you can exclude README.md from the list and other non-essential files.

Required input vars:
- GIT_SOURCE_FILES: The list of files to take the git log hash from.

Output vars:
- GIT_LOG_HASH: The hash of the most recent commit of the files.

## env/git_directory_branch_hash

This sets the NAMESPACE and SUBDOMAIN output variables based on the branch name.

It is hashed to make it url friendly.

Required input vars:
- GIT_BRANCH: The branch name.
- GIT_DIR: The root directory of the git project to hash.

Output vars:
- NAMESPACE: The namespace for the kubernetes deployment.
- SUBDOMAIN: The subdomain for the kubernetes deployment.

## env/git_branch_ticket_number

If your branch name is in the format `ORGCODE-1234/description`, this will extract the ticket number.

Required input vars:
- GIT_BRANCH: The branch name.
- GIT_DIR: The root directory of the git project to hash.

Output vars:
- NAMESPACE: The namespace for the kubernetes deployment. Set to the ticket number.
- SUBDOMAIN: The subdomain for the kubernetes deployment. Set to the ticket number.

## env/git_branch_release_version

If your release branch is in the format `release/1.0.0`, this will extract the version number.

Required input vars:
- GIT_BRANCH: The branch name.
- GIT_DIR: The root directory of the git project to hash.

```go
releaseVersionRFC1123 := strings.ReplaceAll(releaseVersion, ".", "-")
namespace := fmt.Sprintf("release-%s", releaseVersionRFC1123)
```

Output vars:
- NAMESPACE: The namespace for the kubernetes deployment. Set to releaseVersionRFC1123.
- SUBDOMAIN: The subdomain for the kubernetes deployment. Set to releaseVersionRFC1123.

## env/git_directory_branch_static

Sets the NAMESPACE and SUBDOMAIN output variables based on the branch name.

Eg refs/head/develop becomes just develop.

Required input vars:
- GIT_BRANCH: The branch name.
- GIT_DIR: The root directory of the git project to hash.

Output vars:
- NAMESPACE: The namespace for the kubernetes deployment. Set to the branch name.
- SUBDOMAIN: The subdomain for the kubernetes deployment. Set to the branch name.

## env/git_submodule_commit_hash

This sets the docker tag for an image based on the commit hash of the submodule.

Required input vars:
- DOCKER_REGISTRY
- DIR
- SERVICE
- GIT_DIR

Output vars:
- DOCKER_FULL_TAG: The full docker tag for the image.