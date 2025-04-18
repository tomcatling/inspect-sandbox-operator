# Image URL to use all building/pushing image targets
IMG ?= controller:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: generate
generate: ## Generate code including DeepCopy methods for CRDs
	controller-gen object paths="./api/..." crd webhook

.PHONY: vet
vet: generate ## Run go vet against code
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests
	go test ./... -v

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary
	go build -o bin/manager main.go

.PHONY: run
run: fmt vet ## Run against the configured Kubernetes cluster in ~/.kube/config
	go run ./main.go

.PHONY: docker-build
docker-build: ## Build docker image
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image
	docker push ${IMG}

##@ Deployment

.PHONY: install
install: ## Install CRDs into the K8s cluster specified in ~/.kube/config
	kubectl apply -f config/crd/bases

.PHONY: uninstall
uninstall: ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config
	kubectl delete -f config/crd/bases

.PHONY: deploy
deploy: ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	kubectl apply -f config/deploy

.PHONY: undeploy
undeploy: ## Undeploy controller from the configured Kubernetes cluster in ~/.kube/config
	kubectl delete -f config/deploy

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)