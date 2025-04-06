.PHONY: all

all:
	go build -v ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint -c .golangci.yml run ./...
