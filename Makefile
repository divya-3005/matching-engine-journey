.PHONY: all build run test bench fmt vet clean

all: build

build:
	cd server && go build -v -o bin/server .

run:
	cd server && go run .

test:
	cd server && go test -v -race ./...

bench:
	cd server && go test -bench=. -benchmem -run=^$$ ./...

fmt:
	cd server && gofmt -w .

vet:
	cd server && go vet ./...

clean:
	rm -rf server/bin server/wal.log wal.log
