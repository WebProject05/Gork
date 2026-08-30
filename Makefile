.PHONY: build run test clean

BINARY_NAME=gork

build:
	go build -o $(BINARY_NAME) cmd/gork/main.go

test:
	go test -v ./internal/...

run: build
	./$(BINARY_NAME) -c 10 -d 5s https://example.com

clean:
	go clean
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe