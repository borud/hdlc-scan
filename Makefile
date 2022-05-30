all: vet lint test build

build: hdlc-scan

hdlc-scan:
	@go build -o bin/$@

test:
	@go test ./...

vet:
	@go vet ./...

lint:
	@revive ./...

clean:
	@rm -rf bin
