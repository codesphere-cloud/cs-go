
OPENAPI_DIR = ./api/openapi_client

all: format build

format:
	go fmt ./...

lint: install-build-deps
	golangci-lint run

test:
	# -count=1 to disable caching test results
	go test ./api/... ./cli/... ./pkg/... -count=1

test-int: build
	go test ./int/... -count=1

generate: install-build-deps
	go generate ./...

build:
	go build -C cli -o ../cs

GOBIN ?= $(shell go env GOPATH)/bin
install: build
	mv cs ${GOBIN}/

generate-client:
ifeq (, $(shell which openapi-generator-cli))
	$(error "openapi-generator-cli not found, please install, e.g. using brew. See https://openapi-generator.tech/docs/installation")
endif
	rm -rf ${OPENAPI_DIR}
	openapi-generator-cli generate -g go -o ${OPENAPI_DIR} -i https://codesphere.com/api/docs \
	    --additional-properties=generateInterfaces=true,isGoSubmodule=true,withGoMod=false,packageName=openapi_client,disallowAdditionalPropertiesIfNotSet=false \
	    --type-mappings=integer=int \
	    --template-dir openapi-template \
	    --skip-validate-spec # TODO: remove once the Codesphere openapi spec is fixed
	# Remove all non-go files
	rm -r \
		${OPENAPI_DIR}/.gitignore \
		${OPENAPI_DIR}/.openapi-generator/FILES \
		${OPENAPI_DIR}/.openapi-generator-ignore \
		${OPENAPI_DIR}/.travis.yml \
		${OPENAPI_DIR}/api \
		${OPENAPI_DIR}/docs \
		${OPENAPI_DIR}/git_push.sh \
		${OPENAPI_DIR}/README.md \
		${OPENAPI_DIR}/test
	make generate

release-binaries:
	goreleaser release --verbose --skip=validate --skip=publish --clean
 
generate-api: generate-client format

.PHONY: docs
docs:
	rm -rf docs
	mkdir docs
	go run -ldflags="-X 'github.com/codesphere-cloud/cs-go/pkg/io.binName=cs'" hack/gendocs/main.go 
	cp docs/cs.md docs/README.md

generate-license: generate
	go-licenses report --template .NOTICE.template  ./... > NOTICE
	copywrite headers apply

install-build-deps:
ifeq (, $(shell which mockery))
	go install github.com/vektra/mockery/v3@v3.2.1
endif
ifeq (, $(shell which go-licenses))
	go install github.com/google/go-licenses@v1.6.0
endif
ifeq (, $(shell which copywrite))
	go install github.com/hashicorp/copywrite@v0.22.0
endif
ifeq (, $(shell which golangci-lint))
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.2
endif
