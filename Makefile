.PHONY: lint test vendor clean

export GO111MODULE=on

default: lint test

#pvillalobos: if I try running make lint I simply see a "make: golangci-lint: No such file or directory", it'd be good adding a prerequisites part in the readme where any developer can 
#follow up and install whatever third party tools needed before starting, same happens for yaegi_test. This would be for local testing purposes.
lint:
	golangci-lint run

test:
	go test -v -cover ./...

yaegi_test:
	yaegi test -v .

vendor:
	go mod vendor

clean:
	rm -rf ./vendor