build:
	cd cmd/cs && go build
	mv cmd/cs/cs .

install:
	cd cmd/cs && go install
