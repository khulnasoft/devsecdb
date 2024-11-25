#!/bin/sh
# ===========================================================================
# File: build_devsecdb_aws.sh
# Description: usage: ./build_devsecdb_aws.sh
# ===========================================================================

# exit when any command fails
set -e

cd "$(dirname "$0")/../"
. ./scripts/build_init.sh

echo "Start building Devsecdb docker image ${VERSION}..."

docker build -f ./scripts/Dockerfile.aws \
    --build-arg VERSION="${VERSION}" \
    --build-arg GO_VERSION="$(go version)" \
    --build-arg GIT_COMMIT="$(git rev-parse HEAD)" \
    --build-arg BUILD_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"  \
    --build-arg BUILD_USER="$(id -u -n)" \
    -t khulnasoft/devsecdb .

echo "${GREEN}Completed building Devsecdb docker image ${VERSION}.${NC}"
echo ""
echo "Command to tag and push the image"
echo ""
echo "$ docker tag khulnasoft/devsecdb khulnasoft/devsecdb:${VERSION}; docker push khulnasoft/devsecdb:${VERSION}"
echo ""
echo "Command to start Devsecdb on port 8080"
echo ""
echo "$ docker run --init --name devsecdb --restart always --publish 8080:8080 --volume ~/.devsecdb/data:/var/opt/devsecdb khulnasoft/devsecdb:${VERSION} --data /var/opt/devsecdb --port 8080"
echo ""
