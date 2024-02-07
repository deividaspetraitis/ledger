fmt:
	gofmt -s -w .
lint:
	golangci-lint run --build-tags gorillamux --no-config -E gocritic,gofmt,misspell
