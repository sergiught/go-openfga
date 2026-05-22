.PHONY: test integration

test:
	go test ./...

integration:
	cd test/integration && go test ./... -v
