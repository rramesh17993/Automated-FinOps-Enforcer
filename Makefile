# Makefile for FinOps Enforcer

# Image URL to use for building/pushing image targets
IMG ?= finops-enforcer:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

.PHONY: coverage
coverage: test ## Generate coverage report.
	go tool cover -html=cover.out -o coverage.html

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary.
	go build -o bin/manager cmd/controller/main.go

.PHONY: run
run: fmt vet ## Run controller from your host.
	go run cmd/controller/main.go

.PHONY: docker-build
docker-build: ## Build docker image.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image.
	docker push ${IMG}

##@ Deployment

.PHONY: install-crd
install-crd: ## Install CRDs into the cluster.
	kubectl apply -f config/crd/

.PHONY: uninstall-crd
uninstall-crd: ## Uninstall CRDs from the cluster.
	kubectl delete -f config/crd/

.PHONY: deploy
deploy: ## Deploy controller to the cluster.
	kubectl apply -f config/manager/namespace.yaml
	kubectl apply -f config/crd/
	kubectl apply -f config/rbac/
	kubectl apply -f config/manager/

.PHONY: undeploy
undeploy: ## Undeploy controller from the cluster.
	kubectl delete -f config/manager/ --ignore-not-found=true
	kubectl delete -f config/rbac/ --ignore-not-found=true
	kubectl delete -f config/crd/ --ignore-not-found=true

.PHONY: helm-install
helm-install: ## Install via Helm.
	helm install finops-enforcer deploy/helm/finops-enforcer \
		--namespace finops-system \
		--create-namespace

.PHONY: helm-upgrade
helm-upgrade: ## Upgrade Helm release.
	helm upgrade finops-enforcer deploy/helm/finops-enforcer \
		--namespace finops-system

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall Helm release.
	helm uninstall finops-enforcer --namespace finops-system

##@ Local Development

.PHONY: kind-create
kind-create: ## Create kind cluster for local development.
	kind create cluster --name finops-enforcer

.PHONY: kind-delete
kind-delete: ## Delete kind cluster.
	kind delete cluster --name finops-enforcer

.PHONY: kind-load
kind-load: docker-build ## Load docker image into kind cluster.
	kind load docker-image ${IMG} --name finops-enforcer

##@ Utilities

.PHONY: logs
logs: ## Tail controller logs.
	kubectl logs -n finops-system -l app=finops-enforcer -f

.PHONY: describe-policies
describe-policies: ## Describe all enforcement policies.
	kubectl get enforcementpolicies -n finops-system -o wide

.PHONY: apply-samples
apply-samples: ## Apply sample policies.
	kubectl apply -f config/samples/

.PHONY: clean
clean: ## Clean build artifacts.
	rm -rf bin/ cover.out coverage.html
