KO_LOCAL_REGISTRY := ko.local/kube-bench-adapter
KO_REGISTRY ?= github.com/nirmata/kube-bench-adapter
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
LOCAL_PLATFORM ?= linux/arm64
IMAGE_TAG ?= $(shell git describe)

all: build-local

build-local: fmt vet
	CGO_ENABLED=0 go build -o policyreport .

build: fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o policyreport .

fmt:
	go fmt ./...

vet:
	go vet ./...

.PHONY: ko-build
ko-build: fmt vet
	@KO_DOCKER_REPO=$(KO_REGISTRY) \
	ko build . --bare --tags=$(IMAGE_TAG) --platform=$(PLATFORMS)

.PHONY: ko-test-build
ko-test-build:
	@KO_DOCKER_REPO=$(KO_LOCAL_REGISTRY) \
	ko build . --bare --sbom=none --tags=test --platform=$(LOCAL_PLATFORM)

docker: build
	docker build . -t ghcr.io/nirmata/kube-bench-adapter:v0.2.5

codegen:
	./hack/update-codegen.sh
