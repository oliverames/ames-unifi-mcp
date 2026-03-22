BINARY=ames-unifi-mcp
MODULE=github.com/oliveames/ames-unifi-mcp
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build build-all test lint clean docker research

build:
	go build -ldflags "-s -w" -o $(BINARY) ./cmd/ames-unifi-mcp/

build-all:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o dist/$(BINARY)-darwin-arm64 ./cmd/ames-unifi-mcp/
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY)-darwin-amd64 ./cmd/ames-unifi-mcp/
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o dist/$(BINARY)-linux-arm64 ./cmd/ames-unifi-mcp/
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY)-linux-amd64 ./cmd/ames-unifi-mcp/

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/

docker:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(BINARY):$(VERSION) .

research:
	@echo "Before adding a new tool, check these sources:"
	@echo "  - docs/api-research.md (living reference)"
	@echo "  - https://developer.ui.com/"
	@echo "  - https://ubntwiki.com/products/software/unifi-controller/api"
	@echo "  - https://beez.ly/unifi-apis/"
	@echo "  - https://github.com/ubiquiti-community/go-unifi"
	@echo ""
	@echo "Update docs/api-research.md BEFORE implementing the tool."
