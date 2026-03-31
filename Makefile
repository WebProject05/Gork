.PHONY: build run clean

build:
	go build -o gork cmd/gork/main.go

run: build
	./gork -c 10 -d 5s https://example.com

clean:
	rm -f gork