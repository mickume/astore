#!/bin/bash
# Build container image script for Zot Artifact Store

set -e

# Default values
IMAGE_NAME="${IMAGE_NAME:-zot-artifact-store}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
REGISTRY="${REGISTRY:-quay.io}"
CONTAINER_TOOL="${CONTAINER_TOOL:-podman}"
VERSION="${VERSION:-0.1.0-dev}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building Zot Artifact Store Container${NC}"
echo "Image: ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
echo "Version: ${VERSION}"
echo "Tool: ${CONTAINER_TOOL}"
echo

# Build the image
echo -e "${YELLOW}Building image...${NC}"
${CONTAINER_TOOL} build \
    --build-arg VERSION=${VERSION} \
    -t ${IMAGE_NAME}:${IMAGE_TAG} \
    -t ${IMAGE_NAME}:latest \
    -f deployments/container/Containerfile \
    .

echo -e "${GREEN}Build completed successfully!${NC}"
echo

# Optionally tag for registry
if [ ! -z "${REGISTRY}" ]; then
    echo -e "${YELLOW}Tagging for registry...${NC}"
    ${CONTAINER_TOOL} tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
    ${CONTAINER_TOOL} tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:latest
    echo -e "${GREEN}Tagged for ${REGISTRY}${NC}"
fi

echo
echo -e "${GREEN}Image ready!${NC}"
echo "To run: ${CONTAINER_TOOL} run -p 8080:8080 -v \$(pwd)/config:/zot/config:Z ${IMAGE_NAME}:${IMAGE_TAG}"
