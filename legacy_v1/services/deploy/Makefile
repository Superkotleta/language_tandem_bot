.PHONY: lint clean-unused

lint:
	golangci-lint run ./...

clean-unused:
	go mod tidy
	goimports -w .
	golangci-lint run --fix ./...
