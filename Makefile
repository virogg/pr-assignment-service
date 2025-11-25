.PHONY: lint fmt

lint:
	@echo "Running linter..."
	golangci-lint run --config .golangci.yml ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .