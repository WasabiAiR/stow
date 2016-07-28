.PHONY: test
WORKSPACE = $(shell pwd)

topdir = /tmp/$(pkg)-$(version)

all: container runcontainer
	@true

container:
	docker build --no-cache -t builder-stow build/

runcontainer:
	docker run -v $(WORKSPACE):/mnt/src/github.com/graymeta/stow builder-stow

deps:
	go get github.com/Azure/azure-sdk-for-go/storage
	go get github.com/ncw/swift
	go get github.com/cheekybits/is
	go get github.com/aws/aws-sdk-go

test: clean deps vet
	go test -v ./... | tee tests.out
	go2xunit -fail -input tests.out -output tests.xml

vet:
	go vet ./...

clean:
	rm -f tests.out test.xml
