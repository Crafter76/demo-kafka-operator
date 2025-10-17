.PHONY: test test-env

test-env:
	go test -v ./test/...

test:
	go test -v ./...