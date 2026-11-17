.PHONY: all build run test bench fuzz latency profile-cpu profile-mem fmt vet clean reset

all: build

build:
	cd server && go build -v -o bin/server .

run:
	cd server && go run .

test:
	cd server && go test -v -race ./...

bench:
	cd server && go test -bench=. -benchmem -run=^$$ ./...

fuzz:
	cd server && go test -fuzz=FuzzOrderMatchingInvariants -fuzztime=10s ./engine

latency:
	cd server && go test -v -run TestOrderLatencyDistribution ./engine

profile-cpu:
	cd server && go test -bench=BenchmarkProcessOrder$$ -cpuprofile=cpu.pprof -benchmem -run=^$$ ./engine
	@echo "CPU profile written to server/cpu.pprof"

profile-mem:
	cd server && go test -bench=BenchmarkProcessOrder$$ -memprofile=mem.pprof -benchmem -run=^$$ ./engine
	@echo "Memory profile written to server/mem.pprof"

fmt:
	cd server && gofmt -w .

vet:
	cd server && go vet ./...

clean:
	rm -rf server/bin server/wal.log wal.log server/*.pprof

reset:
	@curl -s -f -X POST http://localhost:8081/simulator/reset 2>/dev/null || curl -s -f -X POST http://localhost:8080/simulator/reset 2>/dev/null || (echo "Error: Server is not running on port 8080 or 8081. Start the server with 'make run' first." && exit 1)
	@echo "Market reset complete: order sequence and liquidity re-seeded from #1"
