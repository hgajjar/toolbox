.PHONY: build

dep: ## Get the dependencies and generate mock files
	go get -v -d
	go generate

build: ## Build the binary file
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o build/${LINUX_AMD64_BINARY}
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o build/${LINUX_ARM64_BINARY}
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o build/${DARWIN_AMD64_BINARY}
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o build/${DARWIN_ARM64_BINARY}

test: ## Run tests with race condition detector
	go test -race -short ./...

test-integration: ## Run integration tests
	(cd integration/rabbitmq/consumer && go build -tags integration)
	go test -short -tags=integration ./integration -v