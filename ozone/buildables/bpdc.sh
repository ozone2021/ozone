#!/bin/bash

set -e

SUBMODULE_NAME=$1
SUBMODULE_IMAGE_FULL_TAG=$2
DOCKERFILE_DIR=$3
DOCKER_REGISTRY=$4
GITLAB_PROJECT_CODE=$5
BUILD_ARGS=${@:6}

echo "Building $1 $2 $3 $4 $5 $BUILD_ARGS"

PARAMS_ERROR_MESSAGE="./$SCRIPT_DIR/buildPushDockerImage.sh <submodule_name> <docker_registry> <gitlab_project_id> <docker_build_args>"

if [ -z $SUBMODULE_NAME ] || [ -z $SUBMODULE_IMAGE_FULL_TAG ] || [ -z $DOCKERFILE_DIR ] || [ -z $DOCKER_REGISTRY ] || [ -z $GITLAB_PROJECT_CODE ]
then
  echo $PARAMS_ERROR_MESSAGE
  exit 1
fi
SCRIPT_DIR=`dirname "$0"`

if [ "$IS_CICD_ENV" = "true" ]
then
  TAG_EXISTS=$(python3 $SCRIPT_DIR/tagExistsOnGitlab.py $SUBMODULE_IMAGE_FULL_TAG $GITLAB_PROJECT_CODE)

  echo "Tag exists $TAG_EXISTS"

  if [ "$TAG_EXISTS" = "true" ]
  then
    echo "Tag $SUBMODULE_IMAGE_FULL_TAG already exists in container registry.."
    exit 0
  fi

fi

echo "Directory: $DOCKERFILE_DIR"

if [ -z "$BUILD_ARGS" ]
then
docker build -t "${SUBMODULE_IMAGE_FULL_TAG}" "${DOCKERFILE_DIR}"
else
PREFILL_BUILD_ARGS="docker build ${BUILD_ARGS[@]} -t ${SUBMODULE_IMAGE_FULL_TAG} ${DOCKERFILE_DIR}"
$PREFILL_BUILD_ARGS
fi

printf "\n\n Docker built $SUBMODULE_NAME as $SUBMODULE_IMAGE_FULL_TAG.... \n\n"
docker push "$SUBMODULE_IMAGE_FULL_TAG"