.PHONY: run test fmt

run:
	go run ./cmd/poster

test:
	go test ./...

fmt:
	gofmt -w .
