
OPENAPI_DIR = ./pkg/api/openapi_client

format:
	go fmt ./...

lint:
	golangci-lint run

test:
	go test ./...

build:
	cd cmd/cs && go build
	mv cmd/cs/cs .

install:
	cd cmd/cs && go install

generate-client:
	rm -rf ${OPENAPI_DIR}
	openapi-generator-cli generate -g go -o ${OPENAPI_DIR} -i https://codesphere.com/api/docs \
	    --additional-properties=isGoSubmodule=true,withGoMod=false,packageName=openapi_client \
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
		

generate: generate-client format
