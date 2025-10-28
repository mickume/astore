#!/bin/bash
# Development environment setup script

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Setting up Zot Artifact Store development environment${NC}"
echo

# Check for required tools
echo -e "${YELLOW}Checking required tools...${NC}"

check_tool() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✓${NC} $1 found"
    else
        echo -e "${YELLOW}✗${NC} $1 not found - install from: $2"
        return 1
    fi
}

MISSING_TOOLS=0

check_tool go "https://golang.org/dl/" || MISSING_TOOLS=$((MISSING_TOOLS + 1))
check_tool podman "https://podman.io/getting-started/installation" || MISSING_TOOLS=$((MISSING_TOOLS + 1))
check_tool make "system package manager" || MISSING_TOOLS=$((MISSING_TOOLS + 1))

if [ $MISSING_TOOLS -gt 0 ]; then
    echo
    echo -e "${YELLOW}Please install missing tools before continuing${NC}"
    exit 1
fi

echo
echo -e "${YELLOW}Creating directory structure...${NC}"
mkdir -p config data logs test/testdata deployments/container

echo -e "${YELLOW}Setting up configuration...${NC}"
if [ ! -f "config/config.yaml" ]; then
    cp config/config.yaml.example config/config.yaml
    echo -e "${GREEN}✓${NC} Created config/config.yaml"
else
    echo -e "${GREEN}✓${NC} config/config.yaml already exists"
fi

echo
echo -e "${YELLOW}Installing Go dependencies...${NC}"
go mod download
echo -e "${GREEN}✓${NC} Dependencies downloaded"

echo
echo -e "${YELLOW}Building binary...${NC}"
make build
echo -e "${GREEN}✓${NC} Binary built successfully"

echo
echo -e "${GREEN}Development environment ready!${NC}"
echo
echo "Next steps:"
echo "  1. Build container: make podman-build"
echo "  2. Run container: make podman-run"
echo "  3. Run locally: make run"
echo "  4. Run tests: make test"
echo
echo "For more commands: make help"
