.PHONY: clean

all: vendor build test

test:
	go test --coverprofile=cover.out

convey:
	goconvey -cover=true -excludedDirs vendor

vendor:
	go mod vendor 

build:
	go build ./...

clean:
	go clean 
	rm -f cover.out
