.PHONY: build test lint vet clean

build:
	go build -o biblium ./cmd/biblium

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

vet:
	go vet ./...

clean:
	rm -f biblium coverage.out
