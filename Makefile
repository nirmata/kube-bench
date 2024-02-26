all: docker

build-local: fmt vet
	CGO_ENABLED=0 go build -o policyreport .

build: fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o policyreport .

fmt:
	go fmt ./...

vet:
	go vet ./...

docker: build
	docker build . -t ghcr.io/nirmata/kube-bench-adapter:v0.2.5

codegen:
	./hack/update-codegen.sh
