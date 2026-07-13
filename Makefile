BINARY_NAME := aerocastd
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -ldflags "\
	-s -w \
	-X main.version=$(VERSION) \
	-X main.commitSHA=$(COMMIT)"

.PHONY: build test bench lint docker clean

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/aerocastd/

test:
	go test -v -race -count=1 ./...

bench:
	go test -bench=. -benchmem -count=3 -timeout=300s ./... 2>&1 | tee bench_results.txt

bench-compare:
	@echo "Run 'make bench' first, then compare with previous results using benchstat:"
	@echo "  go install golang.org/x/perf/cmd/benchstat@latest"
	@echo "  benchstat old_bench.txt new_bench.txt"

lint:
	golangci-lint run ./...

docker:
	docker build -t aerocast:$(VERSION) .

clean:
	rm -rf bin/
	rm -f bench_results.txt

run: build
	./bin/$(BINARY_NAME) -config configs/default.yaml

tune-sysctl:
	sudo sysctl -w net.core.somaxconn=65535
	sudo sysctl -w net.ipv4.tcp_max_syn_backlog=65535
	sudo sysctl -w net.ipv4.ip_local_port_range="1024 65535"
	sudo sysctl -w fs.file-max=2097152
