.PHONY: clean

BINARY_NAME=walking-skeleton

all: generate test build

test:
	ginkgo -r 

build:
	go build -o $(BINARY_NAME) -v ./cmd/...

clean:
	go clean 
	rm -f $(BINARY_NAME)

generate:
	go generate

