.PHONY: test
test:
	go test -race -cpu 5 ./...

