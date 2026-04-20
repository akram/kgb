.PHONY: generate build test fmt vet

CONTROLLER_GEN ?= go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.5

generate:
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

build: generate
	mkdir -p bin
	go build -o bin/kgb-controller ./cmd/kgb-controller
	go build -o bin/kgb-gateway ./cmd/kgb-gateway

test: generate
	go test ./...

fmt:
	go fmt ./...

vet: generate
	go vet ./...
