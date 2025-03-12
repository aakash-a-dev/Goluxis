.PHONY: all build test clean example

all: build test example

build:
	go build ./...

test:
	go test -v ./...

example:
	cd examples/hello && go build -o hello

clean:
	rm -f examples/hello/hello
	go clean

lint:
	golangci-lint run

# Run the example server
run-example: example
	./examples/hello/hello

.DEFAULT_GOAL := all 