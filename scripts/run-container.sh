#!/bin/bash
# Run Zot Artifact Store container locally

set -e

# Default values
IMAGE_NAME="${IMAGE_NAME:-zot-artifact-store}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
CONTAINER_TOOL="${CONTAINER_TOOL:-podman}"
CONTAINER_NAME="${CONTAINER_NAME:-zot-artifact-store}"
PORT="${PORT:-8080}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Running Zot Artifact Store Container${NC}"
echo "Image: ${IMAGE_NAME}:${IMAGE_TAG}"
echo "Port: ${PORT}"
echo

# Ensure config directory exists
mkdir -p config data logs

# Create minimal config if it doesn't exist
if [ ! -f "config/config.yaml" ]; then
    echo -e "${YELLOW}Creating default config file...${NC}"
    cp config/config.yaml.example config/config.yaml
fi

# Stop and remove existing container if it exists
if ${CONTAINER_TOOL} ps -a --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${YELLOW}Stopping existing container...${NC}"
    ${CONTAINER_TOOL} stop ${CONTAINER_NAME} || true
    ${CONTAINER_TOOL} rm ${CONTAINER_NAME} || true
fi

# Run the container
echo -e "${YELLOW}Starting container...${NC}"
${CONTAINER_TOOL} run -d \
    --name ${CONTAINER_NAME} \
    -p ${PORT}:8080 \
    -v $(pwd)/config:/zot/config:Z \
    -v $(pwd)/data:/zot/data:Z \
    -v $(pwd)/logs:/zot/logs:Z \
    ${IMAGE_NAME}:${IMAGE_TAG}

echo
echo -e "${GREEN}Container started successfully!${NC}"
echo "Name: ${CONTAINER_NAME}"
echo "Access: http://localhost:${PORT}"
echo
echo "Useful commands:"
echo "  View logs: ${CONTAINER_TOOL} logs -f ${CONTAINER_NAME}"
echo "  Stop: ${CONTAINER_TOOL} stop ${CONTAINER_NAME}"
echo "  Remove: ${CONTAINER_TOOL} rm ${CONTAINER_NAME}"
