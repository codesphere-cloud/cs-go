
OPENAPI_DIR = ./api/openapi_client

all: format build

format:
	go fmt ./...

lint: install-build-deps
	golangci-lint run

test:
	# -count=1 to disable caching test results
	go test ./... -count=1

generate:
	go generate ./...

build:
	cd cmd/cs && go build
	mv cmd/cs/cs .

install:
	cd cmd/cs && go install

generate-client:
	rm -rf ${OPENAPI_DIR}
	openapi-generator-cli generate -g go -o ${OPENAPI_DIR} -i https://codesphere.com/api/docs \
	    --additional-properties=generateInterfaces=true,isGoSubmodule=true,withGoMod=false,packageName=openapi_client \
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


generate-api: generate-client format

generate-license: install-build-deps
	go-licenses report --template .NOTICE.template  ./... > NOTICE
	copywrite headers apply

install-build-deps:
	go install github.com/vektra/mockery/v3@v3.2.1
	go install github.com/google/go-licenses@v1.6.0
	go install github.com/hashicorp/copywrite@v0.22.0
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.2

