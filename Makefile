.PHONY: help build test clean run docker-build podman-build lint coverage

# Variables
BINARY_NAME=zot-artifact-store
VERSION?=0.1.0-dev
BUILD_DIR=bin
CONTAINER_REGISTRY?=quay.io
CONTAINER_IMAGE?=$(CONTAINER_REGISTRY)/zot-artifact-store
GO=go
PODMAN=podman
DOCKER=docker

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
CGO_ENABLED?=0
BUILD_FLAGS=-tags containers_image_openpgp

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/zot-artifact-store

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

test: ## Run all tests
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(BUILD_FLAGS) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-unit: ## Run unit tests only
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(BUILD_FLAGS) -v -race -short ./...

test-integration: ## Run integration tests
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(BUILD_FLAGS) -v -race -run Integration ./...

test-e2e: ## Run end-to-end tests
	$(GO) test -v -race ./test/e2e/...

coverage: test ## Generate coverage report
	$(GO) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	$(GO) vet ./...

tidy: ## Tidy go modules
	$(GO) mod tidy

podman-build: ## Build container image with Podman
	$(PODMAN) build -t $(CONTAINER_IMAGE):$(VERSION) -f deployments/container/Containerfile .
	$(PODMAN) tag $(CONTAINER_IMAGE):$(VERSION) $(CONTAINER_IMAGE):latest

docker-build: ## Build container image with Docker
	$(DOCKER) build -t $(CONTAINER_IMAGE):$(VERSION) -f deployments/container/Containerfile .
	$(DOCKER) tag $(CONTAINER_IMAGE):$(VERSION) $(CONTAINER_IMAGE):latest

podman-run: podman-build ## Run container with Podman
	$(PODMAN) run --rm -p 8080:8080 -v $(PWD)/config:/config:Z $(CONTAINER_IMAGE):latest

docker-run: docker-build ## Run container with Docker
	$(DOCKER) run --rm -p 8080:8080 -v $(PWD)/config:/config $(CONTAINER_IMAGE):latest

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.txt coverage.html
	$(GO) clean

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod verify

update-deps: ## Update dependencies
	$(GO) get -u ./...
	$(GO) mod tidy

.DEFAULT_GOAL := help
