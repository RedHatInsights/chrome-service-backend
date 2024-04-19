#!/bin/bash

set -exv

IMAGE="quay.io/cloudservices/chrome-service"
IMAGE_TAG=$(git rev-parse --short=7 HEAD)
SECURITY_COMPLIANCE_TAG="sc-$(date +%Y%m%d)-$(git rev-parse --short=7 HEAD)"

if [[ -z "$QUAY_USER" || -z "$QUAY_TOKEN" ]]; then
    echo "QUAY_USER and QUAY_TOKEN must be set"
    exit 1
fi

if [[ -z "$RH_REGISTRY_USER" || -z "$RH_REGISTRY_TOKEN" ]]; then
    echo "RH_REGISTRY_USER and RH_REGISTRY_TOKEN  must be set"
    exit 1
fi

# Create tmp dir to store data in during job run (do NOT store in $WORKSPACE)
export TMP_JOB_DIR=$(mktemp -d -p "$HOME" -t "jenkins-${JOB_NAME}-${BUILD_NUMBER}-XXXXXX")
echo "job tmp dir location: $TMP_JOB_DIR"

function job_cleanup() {
    echo "cleaning up job tmp dir: $TMP_JOB_DIR"
    rm -fr $TMP_JOB_DIR
}

trap job_cleanup EXIT ERR SIGINT SIGTERM

DOCKER_CONF="$TMP_JOB_DIR/.docker"
mkdir -p "$DOCKER_CONF"
DOCKER_CONFIG=$DOCKER_CONF docker login -u="$QUAY_USER" -p="$QUAY_TOKEN" quay.io
DOCKER_CONFIG=$DOCKER_CONF docker login -u="$RH_REGISTRY_USER" -p="$RH_REGISTRY_TOKEN" registry.redhat.io
DOCKER_CONFIG=$DOCKER_CONF docker build -t "${IMAGE}:${IMAGE_TAG}" .

if [[ "$GIT_BRANCH" == "origin/security-compliance" ]]; then
    DOCKER_CONFIG=$DOCKER_CONF docker tag "${IMAGE}:${IMAGE_TAG}" "${IMAGE}:${SECURITY_COMPLIANCE_TAG}"
    DOCKER_CONFIG=$DOCKER_CONF docker push "${IMAGE}:${SECURITY_COMPLIANCE_TAG}"
else
    DOCKER_CONFIG=$DOCKER_CONF docker push "${IMAGE}:${IMAGE_TAG}"
fi
