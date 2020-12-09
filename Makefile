# Current Operator version
VERSION ?= 0.0.1

# Default constants
IMG ?= controller:latest
OUT_DIR ?= ./deploy
OUT_CRD ?= ${OUT_DIR}/crd.yaml
OUT_CONTROLLER ?= ${OUT_DIR}/controller.yaml

create-dir:
	mkdir -p ${OUT_DIR}

# Dependencies
dependency:
	go mod download

install-controller-gen:
	GO111MODULE=on go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0

install-kustomize:
	GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3

all-dependencies: dependency install-controller-gen install-kustomize

# Test
fmt:
	go fmt ./...

vet:
	go vet ./...

test: fmt vet
	go test ./... -coverprofile cover.out

# Generate CRDs
generate-crds:
	${GOBIN}/controller-gen crd:trivialVersions=true rbac:roleName=manager-role webhook paths="./..." output:crd:dir=config/crd/bases

manifests: install-controller-gen generate-crds

# Generate deepcopy functions for CRD
generate-deepcopy:
	${GOBIN}/controller-gen object paths="./..." output:dir=pkg/api/v1alpha0

generate: install-controller-gen generate-deepcopy

# Install CRDs into a cluster
kustomize-crds:
	${GOBIN}/kustomize build config/crd > ${OUT_CRD}

install-crds:
	kubectl apply -f ${OUT_CRD}

delete-crds:
	kubectl delete -f ${OUT_CRD}

install-crds-all: manifests install-kustomize kustomize-crds install-crds

delete-crds-all: manifests install-kustomize kustomize-crds delete-crds

# Deploy controller into a cluster
kustomize-controller:
	cd config/manager && ${GOBIN}/kustomize edit set image controller=${IMG}
	${GOBIN}/kustomize build config/default > ${OUT_CONTROLLER}

deploy-controller:
	kubectl apply -f ${OUT_CONTROLLER}

delete-controller:
	kubectl delete -f ${OUT_CONTROLLER}

deploy-controller-all: generate docker-build docker-push manifests install-kustomize kustomize-controller deploy-controller

# Docker
docker-build:
	docker build . -t ${IMG}

docker-push:
	docker push ${IMG}

# Binary
manager: generate fmt vet
	go build -o bin/manager main.go

run: generate fmt vet manifests
	go run ./main.go
